#pragma once
#include <string>
#include <vector>
#include <thread>
#include <atomic>
#include <cstdint>

namespace hdtn {

// Forward declaration for HDTN's CLA callback interface
struct ClaCallbacks {
    void (*onIngressSegment)(const uint8_t* data, size_t length, uint64_t remoteLtpEngineId, void* context);
    void* context;
};

class KissClaPlugin {
public:
    KissClaPlugin();
    ~KissClaPlugin();

    // Plugin lifecycle
    bool Init(const std::string& devicePath, int baudRate, int kissPort, size_t maxFrameSize, size_t ltpMtu, int retryIntervalSec);
    void SetCallbacks(ClaCallbacks callbacks);
    bool Start();
    void Stop();

    // Egress: HDTN hands us an LTP segment to transmit
    bool SendSegment(const uint8_t* data, size_t length, uint64_t remoteLtpEngineId);

    // Statistics
    uint64_t GetFramesSent() const;
    uint64_t GetFramesReceived() const;
    uint64_t GetDecodeErrors() const;
    uint64_t GetSendErrors() const;

private:
    // KISS framing constants
    static constexpr uint8_t FEND  = 0xC0;
    static constexpr uint8_t FESC  = 0xDB;
    static constexpr uint8_t TFEND = 0xDC;
    static constexpr uint8_t TFESC = 0xDD;

    // KISS encode/decode
    std::vector<uint8_t> KissEncode(const uint8_t* data, size_t len);
    bool KissDecode(const std::vector<uint8_t>& frame, std::vector<uint8_t>& output);

    // Serial port management
    bool OpenSerialPort();
    void CloseSerialPort();
    void ConfigureTermios();

    // Receive thread
    void ReceiveLoop();

    // Configuration
    std::string device_path_;
    int baud_rate_;
    int kiss_port_;
    size_t max_frame_size_;
    size_t ltp_mtu_;
    int retry_interval_sec_;

    // State
    int serial_fd_;
    std::thread rx_thread_;
    std::atomic<bool> running_;
    ClaCallbacks callbacks_;

    // Statistics
    std::atomic<uint64_t> frames_sent_;
    std::atomic<uint64_t> frames_received_;
    std::atomic<uint64_t> errors_decode_;
    std::atomic<uint64_t> errors_send_;
};

} // namespace hdtn
