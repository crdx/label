# label

**label** prints text labels on the Brother PT-P710BT over USB or Bluetooth.

## Installation

```sh
go install crdx.org/label/cmd/label@latest
```

## Usage

```
Usage:
    label print [--chain] <text>
    label preview [-m <mm>] <text>
    label status
```


`print` sends the label to the printer. `--chain` prevents the printer from cutting the label. This is useful to avoid the wasted label square between prints. If you forget to remove the flag on your final print, double tap the power button to make it cut and eject the remainder.

`preview` renders in the terminal instead, using the Kitty graphics protocol. It's sized for the given tape width if passed, or tries to get the current width from a connected printer.

The printer is auto-detected: USB is tried first, then Bluetooth. Note that Bluetooth is only supported on Linux.

The default font is Spleen for its aesthetics when printed on the PT-P710BT, but this is subject to change.

### USB permissions

Raw USB access requires permission to open the device node. Either run it as root, or add a udev rule:

```sh
echo 'SUBSYSTEM=="usb", ATTRS{idVendor}=="04f9", ATTRS{idProduct}=="20af", TAG+="uaccess"' \
    | sudo tee /etc/udev/rules.d/70-ptouch.rules
sudo udevadm control --reload
```

Then disconnect and reconnect the printer.

### Bluetooth

Bluetooth uses Linux RFCOMM sockets and BlueZ, so it is only supported on Linux.

Pair the printer once with `bluetoothctl` (use PIN `0000` if prompted):

```sh
[bluetooth]# scan on
[bluetooth]# pair A4:C1:38:AB:CD:EF
[bluetooth]# trust A4:C1:38:AB:CD:EF
```

## Credits

- [ka2n/ptouchgo](https://github.com/ka2n/ptouchgo) for the driver and raster protocol implementation.
- [robby-cornelissen/pt-p710bt-label-maker](https://github.com/robby-cornelissen/pt-p710bt-label-maker) for the Bluetooth protocol details.

## Contributions

Open an [issue](https://github.com/crdx/label/issues) or send a [pull request](https://github.com/crdx/label/pulls).

## Licence

[MIT](LICENCE).
