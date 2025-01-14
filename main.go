package main

import (
	"flag"
	"github.com/devkemc/fundamentos-golang/common"
	"github.com/devkemc/fundamentos-golang/customers"
	"github.com/devkemc/fundamentos-golang/emails"
	"github.com/devkemc/fundamentos-golang/orders"
	"github.com/devkemc/fundamentos-golang/payments"
	"github.com/devkemc/fundamentos-golang/products"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
)

const (
	portFlag = "port"
)

func main() {

	db, err := sqlx.Connect("mysql", os.Getenv(common.ConnectionString))
	if err != nil {
		log.Fatalln(err)
	}
	if err := createTables(db); err != nil {
		panic(err)
	}
	productRepo := products.NewProductRepositorySqlx(common.NewRepositorySqlx(db))
	productService := products.NewProductServiceV1(productRepo)

	customerRepo := customers.NewCustomerRepositorySqlx(common.NewRepositorySqlx(db))
	customerService := customers.NewCustomerServiceV1(customerRepo)

	paymentRepo := payments.NewPaymentRepositorySqlx(common.NewRepositorySqlx(db))
	paymentService := payments.NewPaymentsServiceSimulator(paymentRepo)

	emailService := emails.NewEmailServiceSimulator()

	orderRepo := orders.NewOrderRepositorySqlx(common.NewRepositorySqlx(db))
	orderServ := orders.NewOrderServiceV1(orderRepo, emailService, paymentService, customerService, productService)
	orderHandler := orders.NewOrderHandler(orderServ)

	port := flag.String(portFlag, "8080", "port to server")
	flag.Parse()

	app := fiber.New()
	apiV1 := app.Group("/api/v1")
	orders.SetupRoutes(apiV1, orderHandler)

	err = app.Listen(":" + *port)
	if err != nil {
		panic(err)
		return
	}
}

func createTables(db *sqlx.DB) error {
	// Create customers table first
	customersTable := `
    CREATE TABLE IF NOT EXISTS customers (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) NOT NULL UNIQUE
    );`

	// Create orders table after customers
	ordersTable := `
    CREATE TABLE IF NOT EXISTS orders (
        id INT AUTO_INCREMENT PRIMARY KEY,
        status ENUM('PENDING', 'CONFIRMED', 'CANCELLED') NOT NULL,
        customer_id INT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(customer_id) REFERENCES customers(id)
    );`

	// Create payments table after orders
	paymentsTable := `
    CREATE TABLE IF NOT EXISTS payments (
        id INT AUTO_INCREMENT PRIMARY KEY,
        amount FLOAT NOT NULL,
        type ENUM('CREDIT') NOT NULL,
        status ENUM('PENDING', 'REJECTED', 'CANCELED', 'FAILED', 'ACCEPTED') NOT NULL,
        order_id INT NOT NULL,
        FOREIGN KEY(order_id) REFERENCES orders(id)
    );`

	// Create items table after orders
	productsTable := `
    CREATE TABLE IF NOT EXISTS products (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description VARCHAR(255) NOT NULL,
		price FLOAT NOT NULL
    );`

	// Create items table after orders
	itemsTable := `
    CREATE TABLE IF NOT EXISTS items (
        id INT AUTO_INCREMENT PRIMARY KEY,
        product_id INT NOT NULL,
        quantity INT NOT NULL,
        order_id INT NOT NULL,
        amount FLOAT NOT NULL,
        FOREIGN KEY(order_id) REFERENCES orders(id),
     	FOREIGN KEY(product_id) REFERENCES products(id)
    );`

	// Execute table creation
	tables := []string{customersTable, ordersTable, paymentsTable, productsTable, itemsTable}
	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
			return err
		}
	}
	log.Println("Tables created successfully.")

	defaultCustomer := `
    INSERT INTO customers (name, email)
    VALUES ('Joaquim Kennedy', 'joaquim.kennedy@example.com')
    ON DUPLICATE KEY UPDATE name = name;`

	_, err := db.Exec(defaultCustomer)
	if err != nil {
		log.Fatalf("Failed to insert default customer: %v", err)
		return err
	}

	log.Println("Tables created and default customer with fixed ID inserted successfully.")

	defaultsProducts := `
	INSERT INTO products (id, name, description, price) VALUES
	 (1, 'iphone 15 pro max', 'iphone 15 pro max 256 gb', 5800.99),
	  (2, 'iphone 15 pro', 'iphone 15 pro max 128 gb', 4800.12)
	ON DUPLICATE KEY UPDATE id = id;
`
	_, err = db.Exec(defaultsProducts)
	if err != nil {
		log.Fatalf("Failed to insert defaults produts: %v", err)
		return err
	}
	log.Println("Tables created and defaults products with fixed ID inserted successfully.")
	return nil
}
