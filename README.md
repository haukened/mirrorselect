# MirrorSelect

MirrorSelect is a command-line tool designed to help users select the best mirror for downloading packages. It tests the speed of various mirrors and suggests the fastest one for your location. Since the removal of `netselect` from the Ubuntu repositories, may users have been forced to use the mirror protocol.  This protocol, while useful, only test latency/priximity using a one-sided view of the world.  This leads to decent, but sub-optimal mirrors selections.

### MirrorSelect performs a series of actions to ensure you get the best mirrors:

1. Mirrors are sourced from the [Launchpad Mirror List](https://launchpad.net/ubuntu/+archivemirrors) instead of the [Ubuntu Mirrors Service](http://mirrors.ubuntu.com/).
    1. The Ubuntu Mirrors service does not return the same mirrors every time its accessed, due to the fact its guessing proximity to your GeoIP location and other factors.  This means that the results may be incomplete or different each time its run.
    1. Launchpad has the most up-to-date and complete information on operational official mirrors.
1. Every mirror (in the selected country) is tested for TCP latency instead of ICMP
    1. TCP Latency tests the amount of time it takes for the web server to complete the HTTP request.
    1. Other tools using ICMP latency only test the time the network takes, not how long it takes for the server to process the request.
1. Once the mirrors are ranked by TCP latency for responsiveness, the top `N` servers are then ranked by download speed.
    1. This ensures that you're actually getting the fastest server, not _just_ the quickest one to respond.
    1. Even servers that respond quickly might have lower download speeds due to congestion, or other factors.

## Features

- Automatically selects the fastest mirror
- GeoIP support for automatic country selection, or manual options for privacy.
- Supports protocol selection (HTTP, HTTPS, any)
- Easy to use command-line interface
- Sane defaults

## Installation

To install MirrorSelect, you can choose one of the following options:

1. Installing with go:
    
    ```
    go install github.com/haukened/mirrorselect
    ```

2. Compiling from source:

    ```
    git clone https://github.com/haukened/mirrorselect
    cd mirrorselect
    go get -u ./...
    go build
    ./mirrorselect -h
    ```

## Usage

MirrorSelect provides several command-line flags to customize its behavior:

```sh
mirrorselect [flags]
```

### Command-Line Flags

- `-a`, `--arch`: Select the CPU architecture (default: current system)
- `-c`, `--country`: Filter mirrors by ISO 3166 Alpha-2 country code (Default: Auto-Detect). Options: ISO-3166 Alpha-2 Country Codes.. `US`, `DE`, etc..
- `-h`, `--help`: Display the help message and exit.
- `-m`, `-max`: The max number of servers to perform download speed tests (Default 5)
- `-p`, `--protocol`: Specify the protocol to use (default: ANY). Options: `http`, `https`, `any`.
- `-r`, `--release`: Specify the ubuntu release (Default: Current). Options: Ubuntu codenames e.g. `noble`, `jammy`, `focal`, etc...
- `-t`, `--timeout`: Set the timeout for each mirror latency test in milliseconds (default: 500ms).
- `-v`, `--verbosity`: Show verbose information (Default: WARN) Options: DEBUG, INFO, WARN, ERROR.

## Examples

Select the fastest mirror using default settings:

```sh
mirrorselect
```

Select the fastest HTTPS mirror in the US:

```sh
mirrorselect --protocol https --country US
```

Test 5 mirrors with a timeout of 10 seconds:

```sh
mirrorselect --num-mirrors 5 --timeout 10
```

## Contributing

Contributions are welcome! Please read the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute to this project.

## License

This project is licensed under the GNU GPL v3 License. See the [LICENSE](LICENSE) file for details.

## Contact

For any questions or suggestions, please open an issue on GitHub or contact the maintainer at your.email@example.com.
