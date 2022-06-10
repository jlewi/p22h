/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jlewi/pkg/backend/pkg/datastore"
	"github.com/jlewi/pkg/backend/pkg/gdocs"
	"github.com/jlewi/pkg/backend/pkg/logging"
	"github.com/jlewi/pkg/backend/pkg/output"
	"github.com/jlewi/pkg/backend/pkg/server"
	kfGcp "github.com/kubeflow/internal-acls/google_groups/pkg/gcp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"net"
	"os"
	"os/user"
	"path"
	"path/filepath"
)

type globalOptions struct {
	level string
	debug bool
}

type runOptions struct {
	staticPath      string
	port            int
	credentialsFile string
	secret          string
}

type gcpOptions struct {
	credentialsFile string
	secret          string
}

var (
	log     logr.Logger
	gOpts   globalOptions
	rOpts   runOptions
	gcpOpts gcpOptions

	scopes = []string{
		docs.DocumentsScope,
		drive.DriveScope,
	}

	rootCmd = &cobra.Command{
		Short: "Run the backend for the feed",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			newLogger, err := logging.InitLogger(gOpts.level, gOpts.debug)
			if err != nil {
				panic(err)
			}
			log = *newLogger
		},
	}

	searchCmd = &cobra.Command{
		Use:   "search",
		Short: "Search Google Drive.",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				helper := getWebFlowLocal()
				if helper == nil {
					return errors.New("Unable to create gcp credential helper")
				}
				ts, err := helper.GetTokenSource(context.Background())

				if err != nil {
					return errors.Wrap(err, "Failed to get token source")
				}

				client := oauth2.NewClient(context.Background(), ts)

				docsClient, err := gdocs.NewClient(client, log)

				if err != nil {
					return errors.Wrapf(err, "Failed to create GDocs client")
				}
				// driveId := ""
				// query := "mimeType = 'application/vnd.google-apps.document' and fullText contains 'feed' "

				query := ""
				// Kubeflow(Public) shared drive
				driveId := "0ALYFjr6-o7l8Uk9PVA"
				corpora := "drive"

				stats := &gdocs.QueryStats{}
				f, err := gdocs.NewStatsBuilder(stats)

				if err != nil {
					errors.Wrapf(err, "Failed to create stats builder function")
				}
				if err := docsClient.Search(query, driveId, corpora, f); err != nil {
					return errors.Wrapf(err, "Failed to search drive")
				}

				fmt.Printf("Results:\n%v", output.PrettyString(stats))
				return nil
			}()

			if err != nil {
				log.Error(err, "Failed to connect to Google Drive")
			}
		},
	}
)

func newRunCmd() *cobra.Command {
	var staticPath string
	var port int
	var dbFile string
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run the server.",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
				if err != nil {
					return errors.Wrapf(err, "Failed to listen on port %v", rOpts.port)
				}
				store, err := datastore.New(dbFile, log)

				if err != nil {
					return err
				}
				s, err := server.NewServer(staticPath, listener, store, log)

				if err != nil {
					return errors.Wrapf(err, "Failed to create new server")
				}

				err = s.StartAndBlock()
				if err != nil {
					return errors.Wrapf(err, "Server exited abnormally")
				}
				return nil
			}()

			if err != nil {
				log.Error(err, "Server exited abnormally")
			}
		},
	}

	runCmd.Flags().StringVarP(&staticPath, "static-path", "d", ".", "Static path to serve")
	runCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to serve on")

	dbDefault := getDbDefault()
	runCmd.Flags().StringVarP(&dbFile, "database", "", dbDefault, "The path of the sqllite database to use")

	return runCmd
}

func newFetchDocCmd() *cobra.Command {
	var docId string
	var output string
	var format string
	cmd := &cobra.Command{
		Use:   "getdoc",
		Short: "Get a document from google drive.",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				if format != "json" {
					return errors.Errorf("Invalid value --format=%v; only json is allowed", format)
				}

				helper := getWebFlowLocal()
				if helper == nil {
					return errors.New("Unable to create gcp credential helper")
				}
				ts, err := helper.GetTokenSource(context.Background())

				if err != nil {
					return errors.Wrap(err, "Failed to get token source")
				}

				client := oauth2.NewClient(context.Background(), ts)

				srv, err := docs.NewService(context.Background(), option.WithHTTPClient(client))
				if err != nil {
					return errors.Wrapf(err, "Failed to create docs client service")
				}

				d, err := srv.Documents.Get(docId).Do()
				if err != nil {
					return errors.Wrapf(err, "Failed to get document")
				}

				writer := os.Stdout

				if output != "" {
					writer, err = os.Create(output)
					if err != nil {
						return errors.Wrapf(err, "Could not create file: %v", output)
					}
					defer writer.Close()
				}

				b, err := json.MarshalIndent(d, "", "  ")
				if err != nil {
					return errors.Wrapf(err, "failed to marshal google doc")
				}
				writer.Write(b)
				return nil
			}()

			if err != nil {
				log.Error(err, "Failed to connect to Google Drive")
			}
		},
	}

	cmd.Flags().StringVarP(&gcpOpts.credentialsFile, "credentials-file", "", "", "JSON File containing OAuth2Client credentials as downloaded from APIConsole. Can be a GCS file.")
	cmd.Flags().StringVarP(&docId, "doc", "d", "", "The ID of the doc to fetch")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "The format of the output.")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Optional the file to write to. If not supplied will write to stdout")

	cmd.MarkFlagRequired("doc")
	return cmd
}

func newIndexCmd() *cobra.Command {
	var dbFile string
	var drive string
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index Google Drive.",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				// Create gdocs client
				helper := getWebFlowLocal()
				if helper == nil {
					return errors.New("Unable to create gcp credential helper")
				}
				ts, err := helper.GetTokenSource(context.Background())

				if err != nil {
					return errors.Wrap(err, "Failed to get token source")
				}

				client := oauth2.NewClient(context.Background(), ts)

				gClient, err := gdocs.NewClient(client, log)
				if err != nil {
					return errors.Wrapf(err, "Failed to create docs client service")
				}

				store, err := datastore.New(dbFile, log)

				if err != nil {
					return err
				}

				docsService, err := docs.NewService(context.Background(), option.WithHTTPClient(client))
				if err != nil {
					return errors.Wrapf(err, "failed to create docs service; error %v")
				}
				indexer, err := gdocs.NewIndexer(gClient, docsService, store, log)

				if err != nil {
					return errors.Wrapf(err, "Failed to create drive indexer")
				}

				if err := indexer.Index(drive); err != nil {
					return errors.Wrapf(err, "Failed to index drive %v", drive)
				}
				return nil
			}()

			if err != nil {
				log.Error(err, "Failed to index Google Drive")
			}
		},
	}

	cmd.Flags().StringVarP(&gcpOpts.credentialsFile, "credentials-file", "", "", "JSON File containing OAuth2Client credentials as downloaded from APIConsole. Can be a GCS file.")

	dbDefault := getDbDefault()
	cmd.Flags().StringVarP(&dbFile, "database", "", dbDefault, "The path of the sqllite database to use")

	cmd.Flags().StringVarP(&drive, "drive", "d", "", "The ID of the drive to index")
	cmd.MarkFlagRequired("doc")
	return cmd
}

func getDbDefault() string {
	user, err := user.Current()
	if err != nil {
		fmt.Printf("Failed to get homeDirectory; error %v", err)
		return ""
	}
	homeDirectory := user.HomeDir
	return path.Join(homeDirectory, ".feed", "database.db")
}

func init() {
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(newFetchDocCmd())
	rootCmd.AddCommand(newIndexCmd())

	rootCmd.PersistentFlags().StringVarP(&gOpts.level, "level", "", "info", "The logging level.")
	rootCmd.PersistentFlags().BoolVarP(&gOpts.debug, "debug", "", false, "Enable debug mode for logs.")

	searchCmd.Flags().StringVarP(&gcpOpts.credentialsFile, "credentials-file", "", "", "JSON File containing OAuth2Client credentials as downloaded from APIConsole. Can be a GCS file.")
	searchCmd.Flags().StringVarP(&gcpOpts.secret, "secret", "", "", "The name of a secret in GCP secret manager where the OAuth2 token should be cached. Should be in the form {project}/{secret}")
}

func main() {
	rootCmd.Execute()
}

// getWebFlowLocal returns a credential helper that will cache the credential in a
// local file.
func getWebFlowLocal() *kfGcp.CachedCredentialHelper {
	if gcpOpts.credentialsFile == "" {
		log.Error(errors.New("credentials-file is required"), "credentials-file is required")
		return nil
	}

	webFlow, err := kfGcp.NewWebFlowHelper(gcpOpts.credentialsFile, scopes)

	if err != nil {
		log.Error(err, "Failed to create a WebFlowHelper credential helper")
		return nil
	}

	home, err := os.UserHomeDir()

	if err != nil {
		log.Error(err, "Could not get home directory")
		return nil
	}

	cacheFile := filepath.Join(home, ".cache", "feed", "drive.token")
	h := &kfGcp.CachedCredentialHelper{
		CredentialHelper: webFlow,
		TokenCache: &kfGcp.FileTokenCache{
			CacheFile: cacheFile,
			Log:       log,
		},
		Log: log,
	}

	return h
}
