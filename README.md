# HydraReminder

A minimalist, frictionless tray application that reminds you to stand up or drink water. Runs on **Windows** and **Linux**.

## Features
- **Zero Distractions**: No popups, no sounds, no modal windows. Alerts use a simple red icon and optional blinking.
- **Minimal Footprint**: Lightweight tray application with ~0% CPU and <15MB RAM usage.
- **Smart UI**: Right-click the tray icon to check the remaining time, configure durations, and manage hotkeys natively.
- **Global Hotkey**: Press `CTRL + ALT + <Key>` to instantly reset your active timer from anywhere.
- **Auto-Start**: Integrates with the OS to start on boot (Windows Registry / Linux XDG autostart).

## Platform Support

| Feature              | Windows       | Linux (X11)       |
|----------------------|---------------|-------------------|
| Tray Icon & Menu     | ✅ Native      | ✅ via libayatana  |
| Global Hotkeys       | ✅ Win32 API   | ✅ via X11 (libX11)|
| Autostart            | ✅ Registry    | ✅ XDG `.desktop`  |
| Auto-reset on menu   | ✅             | ❌ (use Reset btn) |

> **Note:** Linux support requires X11. Wayland-only environments are not currently supported for global hotkeys.
> The distro doesn't matter — any Linux desktop with a system tray (GNOME, KDE, XFCE, etc.) will work.

## Making the Icon Visible

### Windows 11
1. Open **Settings** > **Personalization** > **Taskbar**.
2. Expand **Other system tray icons**.
3. Toggle **HydraReminder** to **On**.

### Windows 10
1. Open **Settings** > **Personalization** > **Taskbar**.
2. Click **Select which icons appear on the taskbar**.
3. Toggle **HydraReminder** to **On**.

### Linux (GNOME)
GNOME hides tray icons by default. Install the [AppIndicator extension](https://extensions.gnome.org/extension/615/appindicator-support/) to see tray icons.

## Developer Build Requirements

- **Go 1.22+**

### Windows Build

```bash
go mod tidy
go build -ldflags="-H=windowsgui" -o hydra-reminder.exe ./cmd/hydra-reminder
```

### Linux Build

Install dependencies first:
```bash
sudo apt install -y gcc libgtk-3-dev libayatana-appindicator3-dev libx11-dev
```

Then build:
```bash
go mod tidy
go build -o hydra-reminder ./cmd/hydra-reminder
```

## Open Source Dependencies & Licenses

- **[github.com/getlantern/systray](https://github.com/getlantern/systray)**: Apache License 2.0. Cross-platform tray icon and menu.
- **[golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys)**: BSD 3-Clause. Windows API access.
- **[golang.design/x/hotkey](https://pkg.go.dev/golang.design/x/hotkey)**: MIT. Cross-platform global hotkeys (Linux/macOS).
- Several indirect supporting libraries from the `getlantern` ecosystem under MIT / Apache 2.0.
