# k8s-volume-mount

A tool to mount Kubernetes PersistentVolumeClaims (PVCs) to your local filesystem. This allows you to access data stored in Kubernetes volumes directly from your local machine.

## Features

- Mount Kubernetes PVCs to your local filesystem
- Support for multiple storage providers:
  - WebDAV
  - NFS
  - SFTP
- Automatic port forwarding
- Easy to use command-line interface

## Installation
Download the latest binary for your OS and move it to your PATH.

### Prerequisites for usage

- kubectl configured with access to your Kubernetes cluster
- Mount libraries installed (depending on the provider you want to use)

### Building from source
Go 1.24 or higher is required for building.

```bash
git clone https://github.com/yourusername/k8s-volume-mount.git
cd k8s-volume-mount
go build -o k8s-volume-mount
```

### Installing dependencies
In general, [rclone](https://rclone.org/) is recommended as it is also used for the server pod.
Standard mount tools are also supported.

#### macOS
NFS support is built-in. 
For other protocols, rclone is recommended.  
**Don't install rclone via homebrew because the release does not include the ``rclone mount`` functionality.**  
(see [official install instructions](https://rclone.org/install/))


#### Ubuntu/Debian
```bash
# For WebDAV support
sudo apt-get install davfs2 # or rclone

# For NFS support
sudo apt-get install nfs-common

# For SFTP support
# use rclone
```

## Usage
### Mount a PVC
```bash
k8s-volume-mount mount -pvc my-pvc -namespace my-namespace -provider webdav
```
Options:
 - ``pvc``: Name of the PersistentVolumeClaim to mount (required)
 - ``port``: Specific port for local port forwarding (optional, default: auto-detect)
 - ``provider``: Provider type to use (optional, default: webdav)
   - Available types: webdav, nfs, sftp
 - ``namespace``: k8s namespace
 - ``pause-on-error`` Wait for user input on error before cleanup (allows debugging)
 - ``mount-dir`` Mount directory (optional, default: ~/k8s-mounts)

### Unmount a PVC
```bash
k8s-volume-mount cleanup -pvc my-pvc
```

### Forward remote port to local machine without mounting
This is useful for cases where you want to manually sync or mount.  
Example:
```bash
$ k8s-volume-mount forward -pvc my-pvc -namespace my-namespace -provider webdav
```
```
Creating webdav provider for PVC my-pvc...
Waiting for webdav server for my-pvc to be ready...
Starting port forwarding on port 10000...
LocalPort 10000 is reachable.
Volume my-pvc available at port 10000 via webdav server
k8s-volume-mount config file: /tmp/k8s-volume-mount/my-pvc/config.json
rclone config file: /tmp/k8s-volume-mount/my-pvc/rclone.conf
```
```bash
rclone sync --config /tmp/k8s-volume-mount/my-pvc/rclone.conf srcDir webdav:/destDir
k8s-volume-mount cleanup -pvc my-pvc
```

### List mounted PVCs
```bash
k8s-volume-mount list
```

## How it works

1. The tool creates a temporary deployment in your Kubernetes cluster that mounts the specified PVC
2. Depending on the provider type, it starts a server (WebDAV, NFS, SFTP) in the pod
3. It uses ``kubectl port-forward`` to set up port forwarding from your local machine to the pod
4. It mounts the remote filesystem to your local machine using the appropriate method

## Configuration

The tool uses the following default directories:
- Temporary files: `/tmp/k8s-volume-mount`
- Mount base directory: `k8s-mounts` in your home directory

These can be overridden using environment variables:
- `K8S_VOLUME_MOUNT_TEMP_DIR`: Override the temporary directory
- `K8S_VOLUME_MOUNT_BASE_DIR`: Override the mount base directory

## Logging
Additional logs are stored in the configured temporary directory.

## Troubleshooting

### Port already in use

If you see an error about a port already being in use, you can specify a different port with the `-port` option.

### Connection issues

Make sure your kubectl is properly configured and you have access to the Kubernetes cluster. You can test this with:

```bash
kubectl get pvc
```

### Write delay and missing files with mount command
Heavy write operations may cause that files aren't synced or synced with a delay up to 1 minute.  
Also see [rclone docs](https://rclone.org/commands/rclone_mount/#vfs-cache-mode-writes).  
You can use ``k8s-volume-mount forward`` with ``rclone sync`` to make write operations reliable.

## License
This project is licensed under the MIT License - see the LICENSE file for details.