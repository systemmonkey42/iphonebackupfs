# iPhoneFS

A golang project to mount an unencrypted iPhone backup. It is based on the [Go FUSE file system library](https://bazil.org/fuse/) and sqlite3.

This has only been tested on linux

# Building and Installation

Use `git clone` to obtain the default branch (develop) and install __go 1.19+__

## Dependencies

The following packages (when building on ubuntu) are needed:
- fuse3
- libfuse3-dev

## Building

Build with the following command:

```
go build .
```

If you have issues building, or simply want a smaller executable, try the following:

```
go build --tags osusergo,netgo -o iphonefs -ldflags "-w -s" .
```

Move the resulting binary to somewhere accessible

```
sudo mv iphonefs /usr/local/bin
```

# Usage

```
iphonefs [-A] [-L] [-d <domain>] <backup folder> <mount point>
```

The default mode will present the camera roll at the root of the mount point.  The is the quickest and simplest way to connect and extract images and movies.

To list the available domains, use the "-L" parameters.

To specify a domain, use the `-d <domain>` option. For example, to mount the SMS application and get access to received attachments folder, use the domain `MediaDomain` and browse to the `Library/SMS/Attachments` folder.

To mount the entire backup, use `-A`.  The will cause the domain names to become part of the filesystem.


```
iphonefs /path/to/directory/containing/backup  /mnt/path
```

Note that the backup directory should contain a file called "Metadata.db".

To __dismount__, unmount the destination folder with

```
umount /mnt/path

```

## Environment Variables

The following environment variables are used when starting the application

Name|Description
---|---
ROOT|Root of the backup folder (directory containing manifest.db
MOUNT|Directory to use as mountpoint

Note: To use `$ROOT` but specify a mount point on the command line, specify the empty string `''` as the backup folder.
For example:
```
iphonefs '' /mnt/data
```


# Issues

- All files are readonly
- The iphone metadata is read once prior to making the entire filesystem available, and is never referenced again.
- Metadata inside the backup is ignored.  File timestamps default to current time.
- All backup files are classified into "domains".  By default, only the "CameraRollDomain" is mounted.
- iPhone applications make use of sqlite databases, however opening a sqlite database on a read-only filesystem requires the alternate "url" format with the __immutable__ option set (eg: `file://path/to/sqllite.db?immutable=1`)

