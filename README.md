# MCP Adapter

> 🚀 Transform any HTTP API into MCP (Model Context Protocol) tools, enabling AI assistants to call your APIs

A lightweight HTTP API management and adaptation system that allows you to configure APIs through a visual interface, automatically generate MCP tool definitions, and enable AI assistants like Claude Desktop to directly call your HTTP endpoints.

## ✨ Key Features

- 🎯 **Zero-Code Configuration** - Configure APIs through Web UI without writing code
- 🔌 **MCP Protocol Support** - Automatically convert HTTP APIs to MCP tools
- 🎨 **Custom Type System** - TypeScript-like system for defining reusable complex data structures
- 📦 **Multi-Application Management** - Support for managing multiple independent API applications
- 🌐 **Modern UI** - Responsive design with intuitive operations

## 🚀 Quick Start

### Using Docker (Recommended)

One-click start with no dependencies required:

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  --name mcp-adapter \
  ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

Windows PowerShell:
```powershell
docker run -d -p 8080:8080 -v ${PWD}/data:/app/data --name mcp-adapter ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

After startup, visit: **http://localhost:8080**

### Running from Source

```bash
# Clone the project
git clone https://github.com/yourusername/mcp-adapter.git
cd mcp-adapter

# Install dependencies
go mod download

# Start the service
go run main.go
```

The service will start at `http://localhost:8080`.

## 📖 Usage Workflow

### 1️⃣ Create an Application

Create a new application in the Web interface, for example "Weather API".

### 2️⃣ Define Custom Types (Optional)

If your API uses complex data structures, you can define custom types first.

### 3️⃣ Configure API Endpoints

Add your HTTP API endpoint configuration:

- **Endpoint Name**: GetWeather
- **URL**: https://api.weather.com/current
- **Method**: GET
- **Parameters**: 
  - city (string, query, required)
  - units (string, query, optional)

### 4️⃣ Connect to AI Assistant

Configure Claude Desktop or other MCP clients to connect to:
```
http://localhost:8080/mcp/your-app-path
```

Now your AI assistant can call the configured APIs!

## 🎯 Use Cases

- 🤖 **AI Assistant Enhancement** - Enable Claude and other AI assistants to call your internal APIs
- 🔗 **API Aggregation** - Unified management and invocation of multiple APIs
- 📝 **API Documentation** - Visual management and display of API definitions
- 🧪 **Rapid Prototyping** - Quick configuration and testing of API integrations


## 🔧 Configuration

### Environment Variables

- `PORT` - Service port (default: 8080)
- `DB_TYPE` - Database type: `sqlite` or `mysql` (default: sqlite)
- `DB_PATH` - SQLite database file path (default: ./data/mcp-adapter.db)
- `DB_DSN` - MySQL connection string (e.g.: `user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True`)

### Database Support

Supports both **SQLite** and **MySQL** databases:

#### 🗄️ SQLite (Default)

Zero configuration, ready to use, suitable for small to medium scale:

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

**Features**:
- ✅ Zero configuration, ready to use
- ✅ Lightweight, suitable for individuals and small teams
- ✅ Data persistence, survives restarts
- ✅ Full SQL functionality support

#### 🐬 MySQL

Suitable for production environments and large-scale usage:

```bash
docker run -d \
  -p 8080:8080 \
  -e DB_TYPE=mysql \
  -e DB_DSN="user:password@tcp(mysql-host:3306)/mcp_adapter?charset=utf8mb4&parseTime=True" \
  ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

**Features**:
- ✅ High performance, supports large-scale concurrency
- ✅ Suitable for production environments and cluster deployments
- ✅ Supports master-slave replication and high availability
- ✅ Better data security and backup capabilities

### Docker Data Persistence

**SQLite Mode**: Use volume mount to save data

```bash
docker run -d \
  -p 8080:8080 \
  -v /your/local/path:/app/data \
  ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

**MySQL Mode**: Data is stored in MySQL server, no local directory mount needed

## 🛠️ Tech Stack

- Go + Gin - Backend service
- SQLite - Data storage
- Vanilla JavaScript - Frontend interface
- MCP Protocol - AI assistant protocol

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📝 License

MIT License

