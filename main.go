package main

import (
	"flag"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

type Link struct {
	gorm.Model

	Short string `gorm:"uniqueIndex" json:"short"`
	Long  string `json:"long"`
}

type RPCMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
}

var (
	DatabasePath  = flag.String("database", "shawty.db", "path to the database file")
	AdminUser     = flag.String("admin-user", "admin", "username for the admin RPC")
	AdminPassword = flag.String("admin-password", "admin", "password for the admin RPC")
)

func main() {
	flag.Parse()

	var AdminAuth = basicauth.New(basicauth.Config{Users: map[string]string{*AdminUser: *AdminPassword}})

	db, err := gorm.Open(sqlite.Open(*DatabasePath), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.AutoMigrate(&Link{}); err != nil {
		log.Fatalln(err)
	}

	app := fiber.New()

	app.Get("/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")

		log.Printf("resolving %s...", key)

		var link Link
		if err := db.Where("short = ?", key).First(&link).Error; err != nil {
			return c.Status(404).SendString("Link not found")
		}

		log.Printf("%s -> %s", link.Short, link.Long)

		return c.Redirect(link.Long, fiber.StatusTemporaryRedirect)
	})

	app.Post("/_admin/rpc", AdminAuth, func(c *fiber.Ctx) error {
		var msg RPCMessage
		if err := c.BodyParser(&msg); err != nil {
			return c.Status(400).SendString("Invalid request")
		}

		switch msg.Method {
		case "get_all":
			links, err := getAllLinks(db)
			if err != nil {
				return c.Status(500).SendString("Internal server error")
			}

			return c.JSON(LinksToDto(links))

		case "add":
			if err := addLink(db, msg.Params[0], msg.Params[1]); err != nil {
				return c.Status(400).SendString(err.Error())
			}

			return c.SendString("OK")

		case "remove":
			if err := removeLink(db, msg.Params[0]); err != nil {
				return c.Status(400).SendString(err.Error())
			}

			return c.SendString("OK")

		default:
			return c.Status(400).SendString("Invalid method")
		}
	})

	if err := app.Listen(":3000"); err != nil {
		log.Fatalln(err)
	}
}

func getAllLinks(db *gorm.DB) ([]Link, error) {
	var links []Link
	if err := db.Find(&links).Error; err != nil {
		return nil, err
	}

	return links, nil
}

func addLink(db *gorm.DB, short, long string) error {
	return db.Create(&Link{Short: short, Long: long}).Error
}

func removeLink(db *gorm.DB, short string) error {
	return db.Where("short = ?", short).Delete(&Link{}).Error
}
