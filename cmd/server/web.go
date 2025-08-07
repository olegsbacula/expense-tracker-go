package main

import (
	"encoding/json"
	"expense-tracker-go/model"
	"log"
	"net/url"
	"strconv"

	"azugo.io/azugo"
	"azugo.io/azugo/middleware"
	"azugo.io/azugo/server"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start web server",
	Long: `Web server is the only thing you need to run,
and it takes care of all the other things for you`,
	RunE: runWeb,
}

// func prettyJSON(data []byte) string {
// 	var pretty bytes.Buffer
// 	if err := json.Indent(&pretty, data, "", "  "); err != nil {
// 		return string(data)
// 	}
// 	return pretty.String()
// }

func runWeb(cmd *cobra.Command, args []string) error {
	app, err := server.New(cmd, server.Options{
		AppName: "Example application",
		AppVer:  Version,
	})
	if err != nil {
		return err
	}

	app.Post("/post", func(ctx *azugo.Context) {
		var request model.Request

		err := json.Unmarshal(ctx.Body.Bytes(), &request)

		if err != nil {
			ctx.StatusCode(fasthttp.StatusBadRequest)
			ctx.ContentType("text/plain")
			ctx.Context().SetBodyString("Failed to unmarshal json.")
			return
		}

		const query = `
			INSERT INTO public.expenses (amount, description, expense_type)
			VALUES($1,$2,$3)
		`

		expenses, err := strconv.Atoi(request.Expenses)

		if err != nil {
			ctx.StatusCode(fasthttp.StatusInternalServerError)
			ctx.ContentType("text/plain")
			ctx.Context().SetBodyString("Failed to translate expenses into integer.")
			return
		}
		res, err := db.Exec(query, expenses, request.Description, request.Type)
		if err != nil {
			log.Printf("Failed to insert into database: %v", err)
			ctx.StatusCode(fasthttp.StatusInternalServerError)
			ctx.ContentType("text/plain")
			ctx.Context().SetBodyString("Failed to insert into database.")
			return
		}
		rows, err := res.RowsAffected()
		if err != nil {
			log.Printf("cannot get RowsAffected: %v", err)
		} else {
			log.Printf("Inserted %d row(s)", rows)
		}
		ctx.StatusCode(fasthttp.StatusCreated)
		ctx.Text("Inserted")
	})

	app.Get("/getAllRecords", func(ctx *azugo.Context) {
		const query = `
			SELECT id,amount,description,expense_type FROM expenses;
		`
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("Failed to SELECT from database: %v", err)
			return
		}
		defer rows.Close()
		var results []model.Request
		for rows.Next() {
			var result model.Request
			if err := rows.Scan(&result.ID, &result.Expenses, &result.Description, &result.Type); err != nil {
				log.Printf("row scan error: %v", err)
				continue
			}
			results = append(results, result)
		}
		ctx.StatusCode(fasthttp.StatusOK)
		ctx.ContentType("application/json")
		ctx.JSON(results)
	})

	app.Delete("/deleteExpense/{id}", func(ctx *azugo.Context) {
		const query = `
			DELETE FROM expenses
         	WHERE id = $1;
		`
		res,err :=url.QueryUnescape(ctx.Params.String("id"))
		if len(res) == 0 || err != nil {
			ctx.NotFound()

			return
		}

		resp, err := db.Exec(query, res)
		if err != nil {
			log.Printf("Failed to delete from database: %v", err)
			ctx.StatusCode(fasthttp.StatusInternalServerError)
			ctx.ContentType("text/plain")
			ctx.Context().SetBodyString("Failed to delete from database")
			return
		}
		n, err := resp.RowsAffected()
		if err != nil {
			log.Printf("cannot get RowsAffected: %v", err)
		}
		if n == 0 {
			ctx.StatusCode(fasthttp.StatusNotFound)
			ctx.Context().SetBodyString("Expense not found")
			return
		}

		ctx.StatusCode(fasthttp.StatusOK)
		ctx.ContentType("text/plain")
		ctx.Context().SetBodyString("Worked just fine.")
	})

	app.Patch("/patch", func(ctx *azugo.Context){

		var request model.Request

		err := json.Unmarshal(ctx.Body.Bytes(), &request)

		if err != nil {
			ctx.StatusCode(fasthttp.StatusBadRequest)
			ctx.ContentType("text/plain")
			ctx.Context().SetBodyString("Failed to unmarshal json.")
			return
		}

		const query = `
		UPDATE expenses
        SET
            amount       = $2,
            description  = $3,
            expense_type = $4
        WHERE id = $1;
		`
		resp, err := db.Exec(query,request.ID, request.Expenses,request.Description,request.Type)
		if err != nil {
			log.Printf("Failed to delete from database: %v", err)
			ctx.StatusCode(fasthttp.StatusInternalServerError)
			ctx.ContentType("text/plain")
			ctx.Context().SetBodyString("Failed to update database")
			return
		}

		n, err := resp.RowsAffected()

		if err != nil {
			log.Printf("cannot get RowsAffected: %v", err)
		}
		if n == 0 {
			ctx.StatusCode(fasthttp.StatusNotFound)
			ctx.Context().SetBodyString("Expense not found")
			return
		}

		ctx.StatusCode(fasthttp.StatusOK)
		ctx.ContentType("text/plain")
		ctx.Context().SetBodyString("Patch worked just fine.")

	})

	corsOpts := app.RouterOptions().CORS
	corsOpts.
		SetOrigins("*").
		SetMethods("GET", "POST", "OPTIONS", "DELETE", "PATCH", "PUT").
		SetHeaders("Content-Type", "Authorization", "Accept")

	app.Use(middleware.CORS(&corsOpts))

	server.Run(app)
	return nil
}

func init() {
	initRootCmd()
	if err := godotenv.Load(); err != nil { //go run ./cmd/server to run app.
		log.Println(".env file not found, loading from environment only")
	}
	RootCmd.AddCommand(webCmd)
}
