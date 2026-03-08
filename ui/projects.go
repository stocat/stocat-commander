package ui

import "path/filepath"

type Command struct {
	Name        string
	Exec        string   // e.g. "make", "docker-compose", "docker"
	Args        []string // e.g. []string{"target"} or []string{"up", "-d", "mysql"}
	Description string
	DirName     string // Optional directory override relative to workspace
	IsSeparator bool   // If true, this is just a visual separator in the list
	Interactive bool   // If true, this will run in the foreground (e.g., less)
}

type Project struct {
	Name        string
	Path        string // Absolute path to the project
	RepoURL     string // Github repository URL
	Description string
	Commands    []Command
}

func GetProjects(workspaceDir string) []Project {
	return []Project{
		{
			Name:        "Global Infrastructure",
			Path:        workspaceDir, // Root dir for global
			RepoURL:     "",
			Description: "Shared Infrastructure (DBs, Redis, Consul)",
			Commands: []Command{
				{Name: "Setup Infra (Clone or Pull)", Exec: "sh", Args: []string{"-c", "echo '인프라를 위해 stocat-auth/gateway 클론 또는 업데이트를 진행하겠습니다...' && echo '\n--- Checking stocat-auth ---' && if [ ! -d stocat-auth ]; then git clone https://github.com/stocat/stocat-auth.git; else cd stocat-auth && git pull origin main && cd ..; fi && echo '\n--- Checking stocat-gateway ---' && if [ ! -d stocat-gateway ]; then git clone https://github.com/stocat/stocat-gateway.git; else cd stocat-gateway && git pull origin main && cd ..; fi"}, Description: "필수 저장소(Auth, Gateway) 확인 및 최신화 (제일 먼저 실행)", DirName: ""},
				{IsSeparator: true},
				{Name: "Start All Infrastructure", Exec: "sh", Args: []string{"-c", "cd stocat-auth && docker-compose up -d && cd ../stocat-gateway && docker-compose up -d"}, Description: "모든 인프라 컨테이너 통합 실행 (Consul, DB, Redis)", DirName: ""},
				{Name: "Stop All Infrastructure", Exec: "sh", Args: []string{"-c", "cd stocat-auth && docker-compose down && cd ../stocat-gateway && docker-compose down"}, Description: "모든 인프라 컨테이너 통합 종료", DirName: ""},
				{IsSeparator: true},
				{Name: "Start MySQL", Exec: "docker-compose", Args: []string{"up", "-d", "mysql"}, Description: "백엔드 데이터베이스 (MySQL 8.0) 실행", DirName: "stocat-auth"},
				{Name: "Start Redis", Exec: "docker-compose", Args: []string{"up", "-d", "redis"}, Description: "캐시 및 pub/sub 브로커 (Redis 7) 실행", DirName: "stocat-auth"},
				{Name: "Start Consul", Exec: "docker-compose", Args: []string{"up", "-d", "consul"}, Description: "서비스 디스커버리 (Consul) 실행", DirName: "stocat-gateway"},
				{Name: "Register Configs", Exec: "make", Args: []string{"config-all"}, Description: "Consul에 JWT 키 및 라우트 설정 등록", DirName: "stocat-gateway"},
				{Name: "Check Status", Exec: "docker", Args: []string{"ps", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}"}, Description: "현재 실행중인 인프라 컨테이너 확인", DirName: ""},
				{Name: "Stop Database & Redis", Exec: "docker-compose", Args: []string{"down"}, Description: "MySQL 및 Redis 컨테이너 종료", DirName: "stocat-auth"},
				{Name: "Stop Consul", Exec: "docker-compose", Args: []string{"down"}, Description: "Consul 컨테이너 종료", DirName: "stocat-gateway"},
			},
		},
		{
			Name:        "Asset Service (stocat-asset)",
			Path:        filepath.Join(workspaceDir, "stocat-asset"),
			RepoURL:     "https://github.com/stocat/stocat-asset.git",
			Description: "Asset Scraper & WebSocket API",
			Commands: []Command{
				{Name: "Setup Project (Clone or Pull)", Exec: "sh", Args: []string{"-c", "echo 'stocat-asset 대상 클론 또는 업데이트를 진행하겠습니다...' && if [ ! -d stocat-asset ]; then git clone https://github.com/stocat/stocat-asset.git; else cd stocat-asset && git pull origin main; fi"}, Description: "저장소 클론 또는 최신 코드 가져오기", DirName: "."},
				{IsSeparator: true},
				{Name: "Start All", Exec: "make", Args: []string{"all"}, Description: "모든 Asset 서비스 백그라운드 실행"},
				{Name: "Start Scraper", Exec: "make", Args: []string{"scraper"}, Description: "Asset 데이터 스크래퍼 실행"},
				{Name: "Start WebSocket API", Exec: "make", Args: []string{"websocket"}, Description: "실시간 데이터 제공 WebSocket API 실행"},
				{Name: "Start Toss Crawler", Exec: "make", Args: []string{"crawler"}, Description: "환율 정보 크롤러 실행"},
				{Name: "View Scraper Logs", Exec: "tail", Args: []string{"-n", "+1", "-f", "logs/asset-scraper.log"}, Description: "스크래퍼 실시간 로그 보기 (전체 이력 + 검색)"},
				{Name: "View WebSocket Logs", Exec: "tail", Args: []string{"-n", "+1", "-f", "logs/asset-websocket-api.log"}, Description: "WebSocket API 실시간 로그 보기 (전체 이력 + 검색)"},
				{Name: "View Crawler Logs", Exec: "tail", Args: []string{"-n", "+1", "-f", "logs/exchange-rate-crawler.log"}, Description: "크롤러 실시간 로그 보기 (전체 이력 + 검색)"},
				{Name: "Stop All", Exec: "make", Args: []string{"stop-all"}, Description: "모든 Asset 관련 서비스 종료"},
				{Name: "Clean Ports", Exec: "make", Args: []string{"clean-ports"}, Description: "사용중인 포트 강제 해제"},
			},
		},
		{
			Name:        "Gateway Service (stocat-gateway)",
			Path:        filepath.Join(workspaceDir, "stocat-gateway"),
			RepoURL:     "https://github.com/stocat/stocat-gateway.git",
			Description: "API Gateway, Catalog, Order Services",
			Commands: []Command{
				{Name: "Setup Project (Clone or Pull)", Exec: "sh", Args: []string{"-c", "echo 'stocat-gateway 대상 클론 또는 업데이트를 진행하겠습니다...' && if [ ! -d stocat-gateway ]; then git clone https://github.com/stocat/stocat-gateway.git; else cd stocat-gateway && git pull origin main; fi"}, Description: "저장소 클론 또는 최신 코드 가져오기", DirName: "."},
				{IsSeparator: true},
				{Name: "Start API Gateway", Exec: "make", Args: []string{"gateway"}, Description: "Spring Cloud API Gateway 실행"},
				{Name: "Start Catalog", Exec: "make", Args: []string{"app-catalog"}, Description: "카탈로그 서비스 실행"},
				{Name: "Start Order", Exec: "make", Args: []string{"app-order"}, Description: "주문 처리 서비스 실행"},
				{Name: "View Gateway Logs", Exec: "tail", Args: []string{"-n", "+1", "-f", "logs/gateway.log"}, Description: "게이트웨이 실시간 로그 보기 (전체 이력 + 검색)"},
				{Name: "View Catalog Logs", Exec: "tail", Args: []string{"-n", "+1", "-f", "logs/services:demo-catalog.log"}, Description: "카탈로그 실시간 로그 보기 (전체 이력 + 검색)"},
				{Name: "View Order Logs", Exec: "tail", Args: []string{"-n", "+1", "-f", "logs/services:demo-order.log"}, Description: "주문 실시간 로그 보기 (전체 이력 + 검색)"},
				{Name: "Stop Gateway", Exec: "make", Args: []string{"stop-gateway"}, Description: "게이트웨이 서비스 종료"},
				{Name: "Stop Catalog", Exec: "make", Args: []string{"stop-catalog"}, Description: "카탈로그 서비스 종료"},
				{Name: "Stop Order", Exec: "make", Args: []string{"stop-order"}, Description: "주문 서비스 종료"},
				{Name: "Stop All", Exec: "make", Args: []string{"stop-all"}, Description: "모든 Gateway 관련 서비스 종료"},
			},
		},
		{
			Name:        "Auth Service (stocat-auth)",
			Path:        filepath.Join(workspaceDir, "stocat-auth"),
			RepoURL:     "https://github.com/stocat/stocat-auth.git",
			Description: "Authentication Service",
			Commands: []Command{
				{Name: "Setup Project (Clone or Pull)", Exec: "sh", Args: []string{"-c", "echo 'stocat-auth 대상 클론 또는 업데이트를 진행하겠습니다...' && if [ ! -d stocat-auth ]; then git clone https://github.com/stocat/stocat-auth.git; else cd stocat-auth && git pull origin main; fi"}, Description: "저장소 클론 또는 최신 코드 가져오기", DirName: "."},
				{IsSeparator: true},
				{Name: "Start Auth API", Exec: "make", Args: []string{"auth-api"}, Description: "인증(JWT) 및 권한 관리 서비스 실행"},
				{Name: "View Auth Logs", Exec: "tail", Args: []string{"-n", "+1", "-f", "logs/auth-api.log"}, Description: "인증 서비스 실시간 로그 보기 (전체 이력 + 검색)"},
				{Name: "Stop Auth API", Exec: "make", Args: []string{"stop-auth-api"}, Description: "인증 서비스 종료"},
			},
		},
	}
}
