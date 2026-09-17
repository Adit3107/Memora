// /internal folder means its for backend only

package config

import "os"

type Config struct{
	Port string
}

func Load() Config{ 
	port := os.Getenv("PORT") // Get the value of the "PORT" environment variable. If it's not set, default to "8080".
	if port == "" {
		port = "8080"
	}

	return Config{
		Port: port,
	}
}
// os is package which enable Go to interact with OS, including reading environment variables. The Load function retrieves the value of the "PORT" environment variable and returns a Config struct with the Port field set to that value (or "8080" if it's not set).

// why two returns? The first return is for the Load function, which returns a Config struct. The second return is for the Config struct itself, which is being returned with the Port field set to the value of the "PORT" environment variable (or "8080" if it's not set). This allows other parts of the application to access the configuration settings easily.