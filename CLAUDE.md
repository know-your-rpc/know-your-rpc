# CLAUDE.md - Know Your RPCs

## Build Commands
- Go: `go build ./...` - Build all packages
- Go: `go test ./... -v` - Run all tests with verbose output
- Go: `go test ./server/queries -v` - Run tests in specific package
- Go: `go test ./server/queries/auth_test.go -v` - Run specific test file
- JS/TS: `cd sdk && npm run build` - Build TypeScript SDK
- JS/TS: `cd dump-rpcs && node main.js` - Run RPC dump script

## Code Style Guidelines
- **Go**: Standard Go style (gofmt compliant)
  - Error handling: Always check errors with `if err != nil`
  - Naming: CamelCase for exported items, camelCase for private
  - Context: Pass context as first parameter to functions
- **TypeScript**: ES2020 target, strict mode enabled
  - Use types for all function parameters and returns
  - Prefer async/await over raw promises
  - Use consistent error handling patterns

## Project Structure
- `/server`: Go backend API server
- `/writer`: Data collection service
- `/common`: Shared Go libraries
- `/sdk`: TypeScript client SDK
- `/dump-rpcs`: Scripts for RPC data collection