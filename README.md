# Terminal Fluid Simulation

A real-time, particle-based fluid simulation that runs entirely in your terminal. This project implements Smoothed
Particle Hydrodynamics (SPH) using Go and tcell for rendering.

## Showcase

![Showcase](https://github.com/null-enjoyer/terminal-fluid-simulation/blob/content/content/showcase-magma.gif)
[Youtube Full Video](https://youtu.be/WzcDjDTCxHY)

## Features

- Real-time physics simulation (SPH)
- Multithreaded solver
- Interactive terminal UI
- Mouse and keyboard support
- Customizable physics parameters (gravity, viscosity, density, etc.)
- Preset management
- Wall drawing and erasing

## Installation

### Option 1: Using Go Install

If you have Go installed, you can install the application directly:

```bash
go install github.com/null-enjoyer/terminal-fluid-simulation@latest
```

Ensure your GOPATH bin directory is in your system PATH to run the command globally.

### Option 2: Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/null-enjoyer/terminal-fluid-simulation.git
   cd terminal-fluid-simulation
   ```

2. Download dependencies:
   ```bash
   go mod tidy
   ```

3. Build the binary:
   ```bash
   go build -o terminal-fluid-simulation
   ```

## Usage

### Basic Launch

Run the application with built-in defaults. Note that saving custom presets is disabled in this mode.

```bash
terminal-fluid-simulation
```

### Launch with Configuration

To enable saving custom presets and persistence, provide a path to a configuration file.

```bash
terminal-fluid-simulation --config settings.json
```

If the specified file (e.g., `settings.json`) does not exist, the application will automatically generate it with
default values.

### Command Line Arguments

- `--config`: Path to the settings JSON file
- `--help`: Show help message

## Controls

### Keyboard Shortcuts

| Key            | Action                                           |
|----------------|--------------------------------------------------|
| **Tab**        | Cycle Mouse Mode (Spawn -> Wall -> Erase)        |
| **Space**      | Spawn fluid at cursor position                   |
| **P**          | Pause / Resume simulation                        |
| **R**          | Reset (Remove all fluid particles)               |
| **C**          | Clear all drawn walls                            |
| **W / S**      | Navigate menu up / down                          |
| **A / D**      | Adjust selected menu value                       |
| **Enter**      | Save current preset (only if config file loaded) |
| **Arrow Keys** | Move cursor (alternative to mouse)               |
| **Q**          | Quit application                                 |
| **Esc**        | Quit application / Cancel text input             |

### Mouse Controls

- **Left Click**: Perform the action of the current mode:
    - **Spawn Mode**: Spawns fluid particles
    - **Wall Mode**: Draws wall
    - **Erase Mode**: Removes wall

## Configuration

When running with the `--config <settings-file-path>` flag, you can save adjusted parameters via the on-screen menu: 
- Select "Save" option
- Type name of preset
- Press Enter

Preset will be saved to specified config file.
