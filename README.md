# nproxy

An HTTP/HTTPS proxy server implemented in Go. Supports everything from simple forward proxies to full-featured MITM (Man-in-the-Middle) proxies.

## Features

- **Basic Proxy Functionality**: HTTP request forwarding
- **MITM Proxy Functionality**: HTTPS traffic interception and modification
- **Certificate Generation**: Dynamic server certificate generation
- **Request/Response Modification**: Header and content rewriting
- **Detailed Logging**: Detailed request/response log output
- **Security Header Addition**: Automatic addition of security headers to responses

## Usage

### Starting as a Mock Server

```bash
# Start mock server for testing
go run app/main.go -mock -addr :9090

# Or use Makefile
make run-mock
```

### Starting as a Basic Proxy Server

```bash
# Run directly with Go
go run app/main.go

# Or use Makefile
make run
```

### Starting as a MITM Proxy Server

```bash
# Start MITM proxy (logging only)
go run app/main.go -mitm -addr :8080

# Start MITM proxy (with request/response modification enabled)
go run app/main.go -mitm -modify -v -addr :8080

# Or use Makefile
make run-mitm
make run-mitm-modify
```

### Command Line Options

- `-addr`: Server address (default: `:8080`)
- `-mitm`: Start as MITM proxy
- `-modify`: Enable request/response modification
- `-mock`: Start as mock server
- `-v`: Output detailed logs

### Running with Docker

```bash
# Basic proxy
make start

# MITM proxy
make mitm

# MITM proxy (with modification enabled)
make mitm-modify
```

## Using MITM Proxy

When using the MITM proxy, follow these steps:

1. **Start the proxy**
   ```bash
   make run-mitm
   ```

2. **Install CA certificate**
   - A CA certificate is generated at `./certs/ca.crt` when the proxy starts
   - Install this certificate in your browser or system's trusted certificate store

3. **Configure browser proxy settings**
   - HTTP proxy: `localhost:8080`
   - HTTPS proxy: `localhost:8080`

### CA Certificate Installation Methods

#### macOS
```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ./certs/ca.crt
```

#### Linux (Ubuntu/Debian)
```bash
sudo cp ./certs/ca.crt /usr/local/share/ca-certificates/nproxy-ca.crt
sudo update-ca-certificates
```

#### Windows
Run in PowerShell with administrator privileges:
```powershell
Import-Certificate -FilePath ".\certs\ca.crt" -CertStoreLocation "Cert:\LocalMachine\Root"
```

## MITM Functionality Details

### Request Modification Examples

- Adding `X-MITM-Proxy: true` header
- User-Agent modification
- Adding special headers to API requests

### Response Modification Examples

- Automatic addition of security headers
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `X-XSS-Protection: 1; mode=block`
- HTML content identification and marking
- Adding custom headers

## Testing

```bash
# Run all tests
make test

# Run tests with detailed logs
make test-verbose

# Run specific tests only
go test ./app/proxy/ -run TestMITMProxy
go test ./app/mock/ -run TestHealthEndpoint
```

## Testing Proxy with Mock Server

The project includes a built-in mock server for testing proxy functionality:

### Quick Demo

```bash
# Run automated demo (basic proxy)
./demo.sh basic

# Run automated demo (MITM proxy)
./demo.sh mitm

# Run automated demo (MITM proxy with modification)
./demo.sh mitm-modify
```

### Manual Testing

1. **Start mock server**
   ```bash
   make run-mock
   # Mock server starts on :9090 with endpoints:
   # GET /health - Health check
   # GET /api/users - Mock users API
   # POST /api/echo - Echo request body
   ```

2. **Start proxy server (in another terminal)**
   ```bash
   # Basic proxy
   make run
   
   # OR MITM proxy with modification
   make run-mitm-modify
   ```

3. **Test proxy with mock server**
   ```bash
   # Health check via proxy
   curl -x localhost:8080 http://localhost:9090/health
   
   # Users API via proxy
   curl -x localhost:8080 http://localhost:9090/api/users
   
   # Echo API via proxy
   curl -x localhost:8080 -X POST \
     -H "Content-Type: application/json" \
     -d '{"test":"data"}' \
     http://localhost:9090/api/echo
   ```

### Mock Server Endpoints

- `GET /health`: Returns server health status
- `GET /api/users`: Returns mock user data
- `POST /api/echo`: Echoes request data back
- `GET /*`: Default handler with available endpoints

## Security Considerations

⚠️ **Important**: Use the MITM proxy for educational and debugging purposes only.

- Intercepting other people's network traffic without permission is illegal
- The developers assume no responsibility for any damage caused by using this tool
- Properly manage generated CA certificates and delete them when no longer needed

## License

This project is released under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Bug reports and feature requests are welcome via Issues or Pull Requests.