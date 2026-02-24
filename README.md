# HydraReminder

A minimalist, frictionless Windows 11/10 tray application that reminds you to stand up or drink water. 

## Features
- **Zero Distractions**: No popups, no sounds, no modal windows. Alerts use a simple red icon and optional blinking.
- **Minimal Footprint**: Pure native Win32 API integration ensuring ~0% CPU and <15MB RAM usage. 
- **Smart UI**: Right-click the tray icon to check the remaining time, configure durations, and manage hotkeys natively.
- **Global Hotkey**: Press `CTRL + ALT + <Key>` to instantly reset your active timer from anywhere in Windows.
- **Auto-Start**: Seamlessly integrates with the Windows Registry for clean bootups without relying on secondary services.

## Making the Icon Visible

To ensure you never miss an alert, you should configure Windows to always show the HydraReminder icon on your taskbar instead of burying it in the overflow menu:

**Windows 11:**
1. Open **Windows Settings**.
2. Navigate to **Personalization** > **Taskbar**.
3. Expand **Other system tray icons**.
4. Find **HydraReminder** and toggle the switch to **On**.

**Windows 10:**
1. Open **Windows Settings**.
2. Navigate to **Personalization** > **Taskbar**.
3. Scroll down and click **Select which icons appear on the taskbar**.
4. Find **HydraReminder** and toggle the switch to **On**.

## Developer Build Requirements

- **Go 1.22+**: Core language.
- **Windows OS / Toolkit**: This project directly interfaces with low-level Win32 APIs (`user32.dll`, `RegisterHotKey`, `PostThreadMessageW`) through `golang.org/x/sys/windows`.

### Building from Source

To compile the application as a pure native UI process (without spawning an attached console or command prompt window):

```bash
go mod tidy
go build -ldflags="-H=windowsgui" -o hydra-reminder.exe ./cmd/hydra-reminder
```

The resulting `hydra-reminder.exe` is completely portable and self-contained.

## Open Source Dependencies & Licenses

This project is made possible thanks to the following open-source libraries:

- **[github.com/getlantern/systray](https://github.com/getlantern/systray)**: Used under the **Apache License 2.0**. Powers the lightweight native tray icon and menu system.
- **[golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys)**: Used under the **BSD 3-Clause License**. Provides direct access to the low-level Windows API (`user32.dll` messages and registry functions) without CGO.
- Several indirect supporting libraries from the `getlantern` ecosystem (such as `getlantern/context`, `getlantern/errors`) are used under their respective open-source licenses (primarily MIT / Apache 2.0).

