package main

import (
	"database/sql"
	"log"
	"os"

	apiConfig "order/src/api/config"
	orderController "order/src/order/infrastructure/controller"
	orderUseCase "order/src/order/application/usecase"
	orderClient "order/src/order/infrastructure/client"
	orderPersistence "order/src/order/infrastructure/persistence"
	sharedConfig "order/src/shared/infrastructure/config"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Driver de PostgreSQL
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// getEnv obtiene una variable de entorno o devuelve un valor por defecto
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	log.Println("🚀 Order Service - HITO 3.0 Bootstrap - Iniciando...")

	// Configurar el router con Gin
	router := gin.New()

	// Agregar middlewares básicos necesarios
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Configurar Prometheus metrics si está habilitado
	prometheusEnabled := os.Getenv("PROMETHEUS_ENABLED")
	log.Printf("PROMETHEUS_ENABLED value: '%s'", prometheusEnabled)

	if prometheusEnabled == "true" {
		log.Println("Registering /metrics endpoint for Order service")
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
		log.Println("/metrics endpoint registered successfully for Order service")
	} else {
		log.Println("Prometheus metrics disabled for Order service")
	}

	// Configurar GZIP y otros middlewares compartidos
	gzipSharedCfg := sharedConfig.DefaultSharedConfig()
	sharedConfig.SetupSharedMiddleware(router, gzipSharedCfg)

	// Obtener configuración de la base de datos de variables de entorno
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "order_db")

	// Crear string de conexión
	connStr := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"
	log.Printf("Intentando conectar a %s", connStr)

	// Conectar a la base de datos (opcional para bootstrap)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("⚠️  Advertencia: Error al conectar a la base de datos: %v", err)
		log.Println("⚠️  Continuando sin DB (solo health check)")
		db = nil
	} else {
		defer db.Close()
		// Comprobar la conexión
		err = db.Ping()
		if err != nil {
			log.Printf("⚠️  Advertencia: Error al verificar la conexión a la base de datos: %v", err)
			log.Println("⚠️  Continuando sin DB (solo health check)")
			db = nil
		} else {
			log.Println("✅ Conexión a la base de datos establecida con éxito")
		}
	}

	// API v1 grupo de rutas
	v1 := router.Group("/api/v1")

	// Configurar el módulo API (health check y documentación)
	apiCfg := apiConfig.DefaultAPIConfig()
	apiCfg.DB = db
	apiCfg.Version = "1.0.0-bootstrap"
	apiConfig.SetupAPIModule(router, v1, apiCfg)

	// Configurar módulo Order
	setupOrderModule(v1, db)

	// Iniciar el servidor
	port := getEnv("PORT", "8080")
	log.Printf("✅ Servidor Order Service iniciado en http://localhost:%s", port)
	log.Printf("✅ Health endpoint: GET http://localhost:%s/health", port)
	log.Printf("✅ Health endpoint: GET http://localhost:%s/api/v1/health", port)
	router.Run(":" + port)
}

// setupOrderModule configura el módulo Order
func setupOrderModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Configurando módulo Order...")

	// Crear cliente de stock-service
	stockClient := orderClient.NewStockClient()

	// Crear repositorio de órdenes
	var orderRepo *orderPersistence.OrderPostgresRepository
	if db != nil {
		orderRepo = orderPersistence.NewOrderPostgresRepository(db)
	}

	// Crear casos de uso
	validateStockUC := orderUseCase.NewValidateStockUseCase(stockClient)
	reserveStockUC := orderUseCase.NewReserveStockUseCase(stockClient)
	releaseStockUC := orderUseCase.NewReleaseStockUseCase(stockClient)
	
	var createOrderUC *orderUseCase.CreateOrderUseCase
	var confirmOrderUC *orderUseCase.ConfirmOrderUseCase
	var cancelOrderUC *orderUseCase.CancelOrderUseCase
	var listOrdersUC *orderUseCase.ListOrdersUseCase
	var getOrderUC *orderUseCase.GetOrderUseCase
	if orderRepo != nil {
		createOrderUC = orderUseCase.NewCreateOrderUseCase(orderRepo)
		confirmOrderUC = orderUseCase.NewConfirmOrderUseCase(orderRepo, stockClient)
		cancelOrderUC = orderUseCase.NewCancelOrderUseCase(orderRepo, stockClient)
		listOrdersUC = orderUseCase.NewListOrdersUseCase(orderRepo)
		getOrderUC = orderUseCase.NewGetOrderUseCase(orderRepo)
	}

	// Crear controlador
	orderCtrl := orderController.NewOrderController(validateStockUC, reserveStockUC, releaseStockUC, createOrderUC, confirmOrderUC, cancelOrderUC, listOrdersUC, getOrderUC)

	// Registrar rutas
	orderCtrl.RegisterRoutes(router)

	log.Println("Módulo Order configurado exitosamente")
}
