# iPhoneFS

A golang project to mount an unencrypted iPhone backup. It is based on the [Go FUSE file system library](https://bazil.org/fuse/) and sqlite3.

This has only been tested on linux

# Building and Installation

Use `git clone` to obtain the default branch (develop) and install __go 1.19+__

Build:

```
go build .
```

Move the resulting binary to somewhere accessible

```
sudo mv iphonefs /usr/local/bin
```

# Usage

```
iphonefs /path/to/directory/containing/backup  /mnt/path
```

Note that the backup directory should contain a file called "Metadata.db".

To __dismount__, unmount the destination folder with

```
umount /mnt/path

```

# Issues

- All files are readonly
- The iphone metadata is read once prior to making the entire filesystem available, and is never referenced again.
- Metadata inside the backup is ignored.  File timestamps default to current time.
- All backup files are classified into "domains".  By default, only the "CameraRollDomain" is mounted.
