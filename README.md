# label

**label** prints text labels on the Brother PT-P710BT over USB or bluetooth.

## Installation

```sh
go install crdx.org/label/cmd/label@latest
```

## Usage

```
Usage:
    label [options] print <text>
    label [options] status
```

The printer is auto-detected: USB is tried first, then bluetooth. Note that bluetooth is only supported on Linux.

The default font is Spleen for its aesthetics when printed on the PT-P710BT, but this is subject to change.

### USB permissions

Raw USB access requires permission to open the device node. Either run it as root, or add a udev rule:

```sh
echo 'SUBSYSTEM=="usb", ATTRS{idVendor}=="04f9", ATTRS{idProduct}=="20af", TAG+="uaccess"' \
    | sudo tee /etc/udev/rules.d/70-ptouch.rules
sudo udevadm control --reload
```

Then disconnect and reconnect the printer.

### bluetooth

bluetooth uses Linux RFCOMM sockets and BlueZ, so it is only supported on Linux.

Pair the printer once with `bluetoothctl` (use PIN `0000` if prompted):

```sh
[bluetooth]# scan on
[bluetooth]# pair A4:C1:38:AB:CD:EF
[bluetooth]# trust A4:C1:38:AB:CD:EF
```

## Credits

- [ka2n/ptouchgo](https://github.com/ka2n/ptouchgo) for the driver and raster protocol implementation.
- [robby-cornelissen/pt-p710bt-label-maker](https://github.com/robby-cornelissen/pt-p710bt-label-maker) for the bluetooth protocol details.

## Contributions

Open an [issue](https://github.com/crdx/label/issues) or send a [pull request](https://github.com/crdx/label/pulls).

## Licence

[MIT](LICENCE).
