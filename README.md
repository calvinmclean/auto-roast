# Auto-Roast

Automate your FreshRoast SR800 coffee roaster with free software, 3D-printed parts, and
affordable electronics. TinyGo firmware interacts with the physical buttons on the coffee
roaster using two small motors without any permanent modifications. A Fyne UI allows easy
control from your laptop, complete with the ability to pre-program and replay previous
roasts.

## Getting Started
To get started with Auto-Roast, you'll need:
- FreshRoast SR800
- Computer with Go and TinyGo installed
- The following electronics and 3D printed parts

## Bill of Materials

| Quantity | Component | Use |
| ---: | --- | --- |
| 1 | XIAO RP2040 | Microcontroller. Chosen for its small size and USB-C port |
| 1 | 28BYJ stepper motor | Precisely rotating the roaster's rotary encoder|
| 1 | ULN2003 | Stepper motor driver |
| 1 | Capacitor | Motor supply filtering. Not sure the best to use here. I used 100uF |
| 1 | MG90S servo motor | Click the roaster's button |
| 2 | M3 x 8 mm screws | Attach the stepper to the base |
| 2 | M3 x 4 mm screws | Stepper attachment to the mount |
| 2 | M3 x 8 mm countersunk screws | Stepper motor base attachment to the roaster. Must be countersunk! |
| 1 | M3 x 4 mm screw | Add the bottom lid for the electronics compartment |

### 3D-Printed Parts

| Quantity | Part | Description |
| ---: | --- | --- |
| 1 | [SR800 controller base](https://github.com/calvinmclean/auto-roast/releases/download/stl-v1/SR800.Controller.Base.stl) | Mounts the assembly to the roaster and encloses electronics |
| 1 | [Lid](https://github.com/calvinmclean/auto-roast/releases/download/stl-v1/Lid.stl) | Covers and protects the controller electronics |
| 1 | [Motor gear](https://github.com/calvinmclean/auto-roast/releases/download/stl-v1/Motor.Gear.stl) | Transfers motion from the stepper motor to an encoder gear |
| 1 | [Encoder gear](https://github.com/calvinmclean/auto-roast/releases/download/stl-v1/Encoder.Gear.stl) | Transfers rotation to the roaster's control encoders |
| 1 | [Servo horn](https://github.com/calvinmclean/auto-roast/releases/download/stl-v1/Servo.Horn.stl) | Enables the servo motor to click a button |
| 1 | [Servo strap](https://github.com/calvinmclean/auto-roast/releases/download/stl-v1/Servo.Strap.stl) | Secures the servo motor to the controller base |
| 1 | [Stepper base](https://github.com/calvinmclean/auto-roast/releases/download/stl-v1/Stepper.Base.stl) | Secures the stepper motor to the controller assembly and aids in alignment |

After printing these parts, careful assembly is required to route wires and fit all electronics in the enclosure. It depends a lot on wire size and has some tricky soldering.

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

### Replay Files

The UI can replay a manually-authored file of commands. Use one command per line;
blank lines and lines beginning with `#` are ignored. Add `WAIT <duration>` between
commands using Go duration syntax, such as `WAIT 30s` or `WAIT 3m12s`.
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
