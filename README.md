# AI-Assistant
🤖 High-performance self-hosted AI assistant built with Go. Features real-time chat, multiple AI models via Ollama, excellent concurrency for multi-user environments, and easy network deployment. No external APIs required - runs entirely on your local network.

## Environment Set-Up

### Install Go using Homebrew
brew install go

### Install Ollama
brew install ollama

### Verify installations
go version
ollama --version

### Initialize Go module
go mod init ai-assistant

### Set Go environment (add to ~/.zshrc or ~/.bash_profile)
echo 'export GOPATH=$HOME/go' >> ~/.zshrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.zshrc
echo 'export PATH=$PATH:$GOBIN' >> ~/.zshrc
source ~/.zshrc

### Initialize and Add Dependencies
go mod init ai-assistant
go get github.com/gin-gonic/gin@latest
go get github.com/joho/godotenv@latest
go get github.com/gorilla/websocket@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/sqlite@latest

### Verify dependencies are installed
go mod tidy
go mod verify

### Check go.mod file was created correctly
cat go.mod

### Start Ollama in background
nohup ollama serve > ollama.log 2>&1 &

### Check if it's running
ps aux | grep ollama

### Pull models you want (one-time setup)
ollama pull llama3.2
ollama pull codellama:7b

### Test Ollama is working
curl http://localhost:11434/api/version

### List available models
ollama list

### Develop your Go app
go run main.go

### Stop it later if needed
pkill ollama 

#### Running and Testing (DEVELOPMENT COMMANDS)
##### Create .env file (optional)
cat > .env << EOF
PORT=8080
OLLAMA_URL=http://localhost:11434
ENVIRONMENT=development
DB_PATH=./conversations.db
EOF

##### Run the application in development mode
go run main.go

##### Alternative: Build and run
go build -o ai-assistant
./ai-assistant

##### Test the API endpoints
curl http://localhost:8080/api/health
curl http://localhost:8080/api/models

##### Open web interface
open http://localhost:8080

##### Hot reload during development (install air first)
go install github.com/cosmtrek/air@latest
air init
air

#### Troubleshooting Commands
##### Check if Ollama is running
ps aux | grep ollama
lsof -i :11434

##### Check Go environment
go env GOPATH
go env GOROOT
go env

##### Clear Go module cache if needed
go clean -modcache

##### Check network connectivity
ping localhost
curl -v http://localhost:11434/api/version

##### View application logs
tail -f ai-assistant.log

##### Kill processes if needed
pkill -f ollama
pkill -f ai-assistant