#include "kiss_cla_plugin.h"

#include <fcntl.h>
#include <termios.h>
#include <unistd.h>
#include <cerrno>
#include <cstring>
#include <chrono>

namespace hdtn {

KissClaPlugin::KissClaPlugin()
    : baud_rate_(9600)
    , kiss_port_(0)
    , max_frame_size_(1600)
    , ltp_mtu_(1500)
    , retry_interval_sec_(5)
    , serial_fd_(-1)
    , running_(false)
    , callbacks_{nullptr, nullptr}
    , frames_sent_(0)
    , frames_received_(0)
    , errors_decode_(0)
    , errors_send_(0)
{
}

KissClaPlugin::~KissClaPlugin()
{
    Stop();
}

bool KissClaPlugin::Init(const std::string& devicePath, int baudRate, int kissPort,
                         size_t maxFrameSize, size_t ltpMtu, int retryIntervalSec)
{
    device_path_ = devicePath;
    baud_rate_ = baudRate;
    kiss_port_ = kissPort;
    max_frame_size_ = maxFrameSize;
    ltp_mtu_ = ltpMtu;
    retry_interval_sec_ = retryIntervalSec;

    // Clamp retry interval to valid range [1, 60]
    if (retry_interval_sec_ < 1) retry_interval_sec_ = 1;
    if (retry_interval_sec_ > 60) retry_interval_sec_ = 60;

    return true;
}

void KissClaPlugin::SetCallbacks(ClaCallbacks callbacks)
{
    callbacks_ = callbacks;
}

bool KissClaPlugin::Start()
{
    if (running_.load()) {
        return false; // Already running
    }

    running_.store(true);
    rx_thread_ = std::thread(&KissClaPlugin::ReceiveLoop, this);
    return true;
}

void KissClaPlugin::Stop()
{
    if (!running_.load()) {
        return;
    }

    running_.store(false);

    if (rx_thread_.joinable()) {
        rx_thread_.join();
    }

    CloseSerialPort();
}

bool KissClaPlugin::SendSegment(const uint8_t* data, size_t length, uint64_t /*remoteLtpEngineId*/)
{
    // Reject segments exceeding LTP MTU
    if (length > ltp_mtu_) {
        errors_send_.fetch_add(1, std::memory_order_relaxed);
        return false;
    }

    if (serial_fd_ < 0) {
        errors_send_.fetch_add(1, std::memory_order_relaxed);
        return false;
    }

    // KISS-encode the segment
    std::vector<uint8_t> frame = KissEncode(data, length);

    // Write to serial port
    ssize_t written = ::write(serial_fd_, frame.data(), frame.size());
    if (written < 0 || static_cast<size_t>(written) != frame.size()) {
        errors_send_.fetch_add(1, std::memory_order_relaxed);
        return false;
    }

    frames_sent_.fetch_add(1, std::memory_order_relaxed);
    return true;
}

uint64_t KissClaPlugin::GetFramesSent() const
{
    return frames_sent_.load(std::memory_order_relaxed);
}

uint64_t KissClaPlugin::GetFramesReceived() const
{
    return frames_received_.load(std::memory_order_relaxed);
}

uint64_t KissClaPlugin::GetDecodeErrors() const
{
    return errors_decode_.load(std::memory_order_relaxed);
}

uint64_t KissClaPlugin::GetSendErrors() const
{
    return errors_send_.load(std::memory_order_relaxed);
}

// --- KISS Encode ---
// Produces: FEND + (kiss_port << 4) + escape(data) + FEND
std::vector<uint8_t> KissClaPlugin::KissEncode(const uint8_t* data, size_t len)
{
    std::vector<uint8_t> frame;
    frame.reserve(len * 2 + 3); // Worst case: every byte escaped + 3 overhead

    frame.push_back(FEND);
    frame.push_back(static_cast<uint8_t>((kiss_port_ & 0x0F) << 4)); // Command byte: port in high nibble

    for (size_t i = 0; i < len; ++i) {
        uint8_t b = data[i];
        switch (b) {
            case FEND:
                frame.push_back(FESC);
                frame.push_back(TFEND);
                break;
            case FESC:
                frame.push_back(FESC);
                frame.push_back(TFESC);
                break;
            default:
                frame.push_back(b);
                break;
        }
    }

    frame.push_back(FEND);
    return frame;
}

// --- KISS Decode ---
// Unescapes data from a KISS frame (between FENDs, after command byte).
// Returns true on success, false on invalid escape sequence.
bool KissClaPlugin::KissDecode(const std::vector<uint8_t>& frame, std::vector<uint8_t>& output)
{
    output.clear();

    if (frame.size() < 3) {
        return false; // Too short: need at least FEND + CMD + FEND
    }

    // Find start of data (skip leading FENDs)
    size_t start = 0;
    while (start < frame.size() && frame[start] == FEND) {
        ++start;
    }

    // Find end of data (skip trailing FENDs)
    size_t end = frame.size() - 1;
    while (end > start && frame[end] == FEND) {
        --end;
    }

    if (start > end) {
        return false; // Empty frame
    }

    // Skip command byte
    size_t dataStart = start + 1;

    // Unescape the data
    bool escaped = false;
    for (size_t i = dataStart; i <= end; ++i) {
        uint8_t b = frame[i];
        if (escaped) {
            switch (b) {
                case TFEND:
                    output.push_back(FEND);
                    break;
                case TFESC:
                    output.push_back(FESC);
                    break;
                default:
                    // Invalid escape sequence
                    return false;
            }
            escaped = false;
        } else if (b == FESC) {
            escaped = true;
        } else {
            output.push_back(b);
        }
    }

    // Trailing FESC without following byte is invalid
    if (escaped) {
        return false;
    }

    return true;
}

// --- Serial Port Management ---

bool KissClaPlugin::OpenSerialPort()
{
    serial_fd_ = ::open(device_path_.c_str(), O_RDWR | O_NOCTTY | O_NONBLOCK);
    if (serial_fd_ < 0) {
        return false;
    }

    // Clear non-blocking after open (we use select/poll for reads)
    int flags = ::fcntl(serial_fd_, F_GETFL, 0);
    ::fcntl(serial_fd_, F_SETFL, flags & ~O_NONBLOCK);

    ConfigureTermios();
    return true;
}

void KissClaPlugin::CloseSerialPort()
{
    if (serial_fd_ >= 0) {
        ::close(serial_fd_);
        serial_fd_ = -1;
    }
}

void KissClaPlugin::ConfigureTermios()
{
    struct termios tty;
    std::memset(&tty, 0, sizeof(tty));

    if (::tcgetattr(serial_fd_, &tty) != 0) {
        return;
    }

    // Map baud rate
    speed_t speed;
    switch (baud_rate_) {
        case 1200:   speed = B1200;   break;
        case 2400:   speed = B2400;   break;
        case 4800:   speed = B4800;   break;
        case 9600:   speed = B9600;   break;
        case 19200:  speed = B19200;  break;
        case 38400:  speed = B38400;  break;
        case 57600:  speed = B57600;  break;
        case 115200: speed = B115200; break;
        default:     speed = B9600;   break;
    }

    ::cfsetispeed(&tty, speed);
    ::cfsetospeed(&tty, speed);

    // 8N1, no flow control
    tty.c_cflag &= ~PARENB;        // No parity
    tty.c_cflag &= ~CSTOPB;        // 1 stop bit
    tty.c_cflag &= ~CSIZE;
    tty.c_cflag |= CS8;            // 8 data bits
    tty.c_cflag &= ~CRTSCTS;       // No hardware flow control
    tty.c_cflag |= CREAD | CLOCAL; // Enable receiver, ignore modem control

    // Raw input mode
    tty.c_lflag &= ~(ICANON | ECHO | ECHOE | ISIG);

    // No software flow control
    tty.c_iflag &= ~(IXON | IXOFF | IXANY);
    tty.c_iflag &= ~(IGNBRK | BRKINT | PARMRK | ISTRIP | INLCR | IGNCR | ICRNL);

    // Raw output
    tty.c_oflag &= ~OPOST;

    // Read with timeout: return after 100ms or 1 byte
    tty.c_cc[VMIN] = 0;
    tty.c_cc[VTIME] = 1; // 100ms timeout

    ::tcsetattr(serial_fd_, TCSANOW, &tty);
}

// --- Receive Loop ---

void KissClaPlugin::ReceiveLoop()
{
    std::vector<uint8_t> frame_buffer;
    bool in_frame = false;

    while (running_.load()) {
        // Ensure serial port is open; retry if not
        if (serial_fd_ < 0) {
            if (!OpenSerialPort()) {
                // Sleep for retry interval then try again
                for (int i = 0; i < retry_interval_sec_ * 10 && running_.load(); ++i) {
                    std::this_thread::sleep_for(std::chrono::milliseconds(100));
                }
                continue;
            }
        }

        // Read one byte at a time
        uint8_t byte;
        ssize_t n = ::read(serial_fd_, &byte, 1);

        if (n < 0) {
            if (errno == EAGAIN || errno == EWOULDBLOCK) {
                continue; // No data available, loop again
            }
            // Read error — close port and retry
            CloseSerialPort();
            continue;
        }

        if (n == 0) {
            // Timeout (VTIME expired with no data), just loop
            continue;
        }

        // Process the byte through KISS state machine
        if (byte == FEND) {
            if (in_frame && !frame_buffer.empty()) {
                // End of frame — decode and deliver
                // Wrap in full KISS frame format for KissDecode
                std::vector<uint8_t> full_frame;
                full_frame.push_back(FEND);
                full_frame.insert(full_frame.end(), frame_buffer.begin(), frame_buffer.end());
                full_frame.push_back(FEND);

                std::vector<uint8_t> decoded;
                if (KissDecode(full_frame, decoded)) {
                    if (!decoded.empty() && callbacks_.onIngressSegment) {
                        callbacks_.onIngressSegment(decoded.data(), decoded.size(), 0, callbacks_.context);
                        frames_received_.fetch_add(1, std::memory_order_relaxed);
                    }
                } else {
                    errors_decode_.fetch_add(1, std::memory_order_relaxed);
                }

                frame_buffer.clear();
                in_frame = false;
            } else {
                // Start of a new frame (or inter-frame FEND)
                frame_buffer.clear();
                in_frame = true;
            }
        } else {
            if (in_frame) {
                frame_buffer.push_back(byte);

                // Check max frame size
                if (frame_buffer.size() > max_frame_size_) {
                    // Frame too large — discard
                    errors_decode_.fetch_add(1, std::memory_order_relaxed);
                    frame_buffer.clear();
                    in_frame = false;
                }
            }
            // If not in_frame, discard bytes (garbage before first FEND)
        }
    }

    CloseSerialPort();
}

} // namespace hdtn
