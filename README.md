# Auto-Roast

Auto-Roast is a software project designed to automate the control and monitoring of coffee roasting. It integrates with hardware components for precision control and supports automation scripts to replicate or fine-tune roast profiles.

## Getting Started
To get started with Auto-Roast, you'll need:
- A compatible hardware setup (e.g., FreshRoast coffee roaster).
- [Go](https://golang.org) installed (version 1.25.1 or later).
- TinyGo for building and flashing the firmware.

## Features
- **Serial Command Automation:** Automate command sequences, including pre-heat, pauses, and dynamic adjustments during roasting.
- **Roast Log Conversion:** Transform roast logs into command lists for replication or refinement.
- **Calibration Tools:** Adjust stepper and servo settings dynamically.
- **Multi-mode Control:** Supports fan, timer, and power adjustment modes.

## Installation
1. Clone the repository:
    ```bash
    git clone https://github.com/calvinmclean/auto-roast.git
    cd auto-roast
    ```
2. Install TinyGo for firmware tasks:
    ```bash
    # Follow TinyGo installation instructions at https://tinygo.org/getting-started/
    ```
3. **Flash Firmware:** Use TinyGo to upload the compiled firmware to your microcontroller. This allows it to interface with the rest of the Auto-Roast system. See `Taskfile.yml` for automation.
    ```bash
    task flash
    ```
4. **Run Companion Program:** Launch the accompanying control program on your computer to send commands and analyze roast data in real-time.
    ```bash
    task run -- -session="El Salvador Dry Process Finca San Luis"
    ```
5. **Direct Interaction:** Use TinyGo's serial tools to communicate directly with the controller for advanced debugging or manual control.
    ```bash
    tinygo flash -monitor -target=pico ./firmware
    ```

## Usage

### Running Commands
Tasks for development, testing, and deployment are managed via the Taskfile. Examples:
- **Run Serial Tests:**
    ```bash
    task serial-test
    ```
- **Build Firmware:**
    ```bash
    task build
    ```
- **Flash Firmware:**
    ```bash
    task flash
    ```
- **Run Application:**
    ```bash
    task run -- <CLI_ARGS>
    ```

### TWChart Integration

Auto-Roast integrates with [TWChart](http://github.com/calvinmclean/twchart), a system that integrates with Thermoworks Cloud thermometers to record temperature data and overlay events and notes. This integration enables visualization of roast profiles, adjustments, and logs for better analysis.

### Prerequisites
- Ensure `TWCHART_ADDR` is correctly set to the address where your TWChart server is running. For example:
    ```bash
    export TWCHART_ADDR=http://localhost:8080
    ```
  Alternatively, you can modify the `Taskfile.yml` to set this address in the `run` task configuration.

### Usage
- During runtime, the Auto-Roast application will continuously send roast data to the configured TWChart instance.
- After roasting is complete, follow TWChart's instructions for uploading temperature data from Thermoworks Cloud
- Historical roast logs can also be analyzed via TWChart for refining roast profiles.

---

## Configuration
Set the following environment variables as needed:
- `TWCHART_ADDR`: Address of the TwinChart server (e.g., `http://localhost:8080`).
- `IGNORE_SERIAL`: Ignore serial interfaces (used for development).
