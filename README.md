# winup

Automate
[downloading](https://developer.microsoft.com/en-us/microsoft-edge/tools/vms/)
and setting up a Windows 10 64-bit VM on VirtualBox.

Still a Work In Progress, so please file bugs.

### Requirements

* VirtualBox v6 with VT-x or alike
* 4 GiB spare RAM recomended
* 50 GiB spare SSD recommended

A decent CPU and internet connection are also recommended, as the process
requires running Windows and downloading many large archives. TODO: add
estimates of final time/disk/net usage.

### Objectives

The final VM snapshot will be better than the original from Microsoft:

* Fast: the live snapshot starts up in ~5s on my laptop
* Usable: SSH access, including login as admin (TODO ssh part)
* Idle: less bloatware, no telemetry, fewer background services

The final image has the users `ieuser` and `administrator`, both with the
default password `Passw0rd!`.

It should also be trivial to install developer tools on the final VM, such as
Git or Go. That's also a TODO for now.

### Notes

The process currently relies on VirtualBox. I considered Vagrant, but did not
want to add more dependencies. VirtualBox seems powerful and portable enough.

The initial VM setup steps happen with its virtual network unplugged, to keep
Windows from installing random software, and to speed up the process.

If you have issues with VirtualBox on Linux, see the [Arch Wiki
page](https://wiki.archlinux.org/index.php/VirtualBox).

### Development

The project is script-like, using panics and helper functions to keep the code
short and readable. Almost all encountered errors stop the process, anyway.

Sleeps should be avoided at all costs. Waits and events are better than periodic
polling, if possible.

Continuous Integration would be good, but I don't know a free service that's
powerful enough. I don't have a desktop or server powerful enough lying around
for the job.

### License

All code in this repository is BSD-3 licensed; see [LICENSE]().

All other pieces of software, including the Windows 10 VM image, are not
distributed with this git repository.
