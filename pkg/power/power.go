// Package power provides ultra-low-power mode management for STM32U585-based nodes.
//
// The STM32U585 supports multiple low-power modes:
// - Stop 2 mode: ~16 µA idle current
// - Standby mode: ~2 µA with RTC
// - Shutdown mode: ~170 nA
//
// This package implements Stop 2 mode with wake-on-contact for scheduled passes.
package power

import (
	"fmt"
	"sync"
	"time"
)

// PowerMode represents the STM32U585 power mode
type PowerMode int

const (
	// PowerModeRun is normal operation mode
	PowerModeRun PowerMode = iota
	// PowerModeStop2 is Stop 2 ultra-low-power mode (~16 µA)
	PowerModeStop2
	// PowerModeStandby is Standby mode with RTC (~2 µA)
	PowerModeStandby
	// PowerModeShutdown is Shutdown mode (~170 nA)
	PowerModeShutdown
)

func (p PowerMode) String() string {
	switch p {
	case PowerModeRun:
		return "Run"
	case PowerModeStop2:
		return "Stop2"
	case PowerModeStandby:
		return "Standby"
	case PowerModeShutdown:
		return "Shutdown"
	default:
		return fmt.Sprintf("Unknown(%d)", p)
	}
}

// PowerManager manages STM32U585 power modes
type PowerManager struct {
	mu            sync.Mutex
	currentMode   PowerMode
	wakeupTime    time.Time
	wakeupEnabled bool
	callbacks     []func()
}

// Config holds power manager configuration
type Config struct {
	// Enable wake-on-contact for scheduled passes
	WakeOnContact bool

	// Minimum time in Stop 2 mode before wakeup (prevents thrashing)
	MinSleepDuration time.Duration
}

// DefaultConfig returns default power manager configuration
func DefaultConfig() Config {
	return Config{
		WakeOnContact:    true,
		MinSleepDuration: 10 * time.Second,
	}
}

// New creates a new power manager
func New(config Config) *PowerManager {
	return &PowerManager{
		currentMode: PowerModeRun,
		callbacks:   make([]func(), 0),
	}
}

// GetMode returns the current power mode
func (p *PowerManager) GetMode() PowerMode {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentMode
}

// EnterStop2 enters Stop 2 ultra-low-power mode
func (p *PowerManager) EnterStop2(wakeupTime time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentMode != PowerModeRun {
		return fmt.Errorf("cannot enter Stop 2 from %s mode", p.currentMode)
	}

	// In real implementation, this would:
	// 1. Save peripheral states
	// 2. Configure RTC wakeup timer
	// 3. Disable unnecessary clocks
	// 4. Enter Stop 2 mode via PWR registers
	// 5. CPU halts, ~16 µA current draw
	// For simulation, we just track the mode and schedule wakeup

	p.currentMode = PowerModeStop2
	p.wakeupTime = wakeupTime
	p.wakeupEnabled = true

	duration := time.Until(wakeupTime)
	fmt.Printf("Power: Entered Stop 2 mode (wakeup in %s at %s)\n",
		duration.Round(time.Second), wakeupTime.Format("15:04:05"))

	// Schedule wakeup
	go p.scheduleWakeup(duration)

	return nil
}

// ExitStop2 exits Stop 2 mode and returns to Run mode
func (p *PowerManager) ExitStop2() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentMode != PowerModeStop2 {
		return fmt.Errorf("not in Stop 2 mode")
	}

	// In real implementation, this would:
	// 1. Restore system clocks
	// 2. Restore peripheral states
	// 3. Re-enable interrupts
	// 4. Resume normal operation
	// For simulation, we just change the mode

	p.currentMode = PowerModeRun
	p.wakeupEnabled = false

	fmt.Println("Power: Exited Stop 2 mode, resumed Run mode")

	// Execute wakeup callbacks
	for _, callback := range p.callbacks {
		callback()
	}

	return nil
}

// scheduleWakeup schedules automatic wakeup from Stop 2 mode
func (p *PowerManager) scheduleWakeup(duration time.Duration) {
	time.Sleep(duration)

	p.mu.Lock()
	if p.wakeupEnabled && p.currentMode == PowerModeStop2 {
		p.mu.Unlock()
		p.ExitStop2()
	} else {
		p.mu.Unlock()
	}
}

// RegisterWakeupCallback registers a callback to be called on wakeup
func (p *PowerManager) RegisterWakeupCallback(callback func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.callbacks = append(p.callbacks, callback)
}

// GetCurrentDraw returns estimated current draw in microamps
func (p *PowerManager) GetCurrentDraw() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.currentMode {
	case PowerModeRun:
		return 50000.0 // ~50 mA typical for STM32U585 at 160 MHz
	case PowerModeStop2:
		return 16.0 // ~16 µA in Stop 2 mode
	case PowerModeStandby:
		return 2.0 // ~2 µA in Standby with RTC
	case PowerModeShutdown:
		return 0.17 // ~170 nA in Shutdown
	default:
		return 0
	}
}

// GetPowerBudget calculates power consumption over a time period
func (p *PowerManager) GetPowerBudget(runTime, sleepTime time.Duration) PowerBudget {
	runCurrent := 50000.0  // µA in Run mode
	sleepCurrent := 16.0   // µA in Stop 2 mode

	totalTime := runTime + sleepTime
	avgCurrent := (runCurrent*runTime.Seconds() + sleepCurrent*sleepTime.Seconds()) / totalTime.Seconds()

	// Voltage is 3.3V for STM32U585
	voltage := 3.3
	avgPower := avgCurrent * voltage / 1e6 // Convert µA to A, then to Watts

	return PowerBudget{
		RunTime:        runTime,
		SleepTime:      sleepTime,
		RunCurrent:     runCurrent,
		SleepCurrent:   sleepCurrent,
		AverageCurrent: avgCurrent,
		AveragePower:   avgPower,
		Voltage:        voltage,
	}
}

// PowerBudget holds power consumption calculations
type PowerBudget struct {
	RunTime        time.Duration
	SleepTime      time.Duration
	RunCurrent     float64 // µA
	SleepCurrent   float64 // µA
	AverageCurrent float64 // µA
	AveragePower   float64 // Watts
	Voltage        float64 // Volts
}

// String returns a string representation of the power budget
func (b PowerBudget) String() string {
	return fmt.Sprintf("Power Budget: Run=%s Sleep=%s AvgCurrent=%.1fµA AvgPower=%.3fW",
		b.RunTime.Round(time.Second),
		b.SleepTime.Round(time.Second),
		b.AverageCurrent,
		b.AveragePower)
}

// IsWithinBudget checks if the power budget is within the specified limit
func (b PowerBudget) IsWithinBudget(maxPowerWatts float64) bool {
	return b.AveragePower <= maxPowerWatts
}

// WakeOnContact configures wakeup for a scheduled contact window
func (p *PowerManager) WakeOnContact(contactStartTime time.Time, prepTime time.Duration) error {
	// Wake up prepTime before the contact window starts
	wakeupTime := contactStartTime.Add(-prepTime)

	if time.Now().After(wakeupTime) {
		return fmt.Errorf("wakeup time is in the past")
	}

	return p.EnterStop2(wakeupTime)
}

// GetTimeUntilWakeup returns the time until scheduled wakeup
func (p *PowerManager) GetTimeUntilWakeup() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.wakeupEnabled {
		return 0
	}

	return time.Until(p.wakeupTime)
}

// IsAsleep returns whether the system is in a low-power mode
func (p *PowerManager) IsAsleep() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentMode != PowerModeRun
}
