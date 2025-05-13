# Start Command

The `start` command allows you to launch a Microcks instance using a container runtime like Docker or Podman. It either starts a new container or resumes a previously created one based on saved configuration.

📌 Usage
```bash
microcks start [flags]
```

🚩 Flags
| Flag       | Description                                        | Required | Default                                        |
| ---------- | -------------------------------------------------- | -------- | ---------------------------------------------- |
| `--name`   | Name of the Microcks container/instance            | No       | `microcks`                                     |
| `--port`   | Host port to expose Microcks                       | No       | `8585`                                         |
| `--image`  | Image to use for creating the container            | No       | `quay.io/microcks/microcks-uber:latest-native` |
| `--rm`     | Auto-remove container on exit (like `docker --rm`) | No       | `false`                                        |
| `--driver` | Container runtime to use (`docker` or `podman`)    | No       | `docker`                                       |


🧪 Examples

Start Microcks with default settings:
```sh
microcks start
```
Start Microcks on port 9090 using Podman:
```sh
microcks start --port 9090 --driver podman
```

Start with a custom container name and image:
```sh
microcks start --name dev-microcks --image custom/microcks:latest
```
