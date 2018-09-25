// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	client "github.com/ichbinjoe/sbts_client/client"
)

var (
	CertLocation       string
	KeyLocation        string
	OverrideCALocation string
	SkipInsecure       bool
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get uses sbts to attempt to retrieve a file from a sbts server",
	Long: `Get takes a URL and attempts to retrieve the file from a remote sbts
	server. Example:
	
	sbts_client get sbts://server/file
	
	If you want to use tls, you can use
	
	sbts_client get sbtss://server/file
	
	There are flags available to set up client and server certificates.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("Invalid usage - expected url as only argment.")
			return
		}

		u, e := url.Parse(args[0])
		if e != nil {
			fmt.Printf("Invalid usage - %s is not a valid url.", args[0])
			return
		}

		// we don't use a dialer, because idk. let someone else do that
		c := client.Config{}

		slc := strings.ToLower(u.Scheme)

		if slc == "sbtss" {
			// we need to do some additional stuff if we are doing tls

			var clientCert []tls.Certificate
			if CertLocation != "" {
				if KeyLocation == "" {
					fmt.Printf("Invalid usage - key not given with cert!")
					return
				}

				c, e := tls.LoadX509KeyPair(CertLocation, KeyLocation)

				if e != nil {
					fmt.Printf("Error while loading X509 key pair: %v", e)
					return
				}

				clientCert = []tls.Certificate{c}
			}

			var capool *x509.CertPool

			if OverrideCALocation != "" {
				f, e := os.Open(OverrideCALocation)
				if e != nil {
					fmt.Printf("Unable to open %s - %v\n", OverrideCALocation, e)
					return
				}

				defer f.Close()

				fi, e := f.Stat()
				if e != nil {
					fmt.Printf("Error while statting %s - %v\n", OverrideCALocation, e)
					return
				}

				capool = x509.NewCertPool()

				loadCA := func(f *os.File) error {
					d, e := ioutil.ReadAll(f)
					if e != nil {
						return e
					}
					c, e := x509.ParseCertificate(d)
					if e != nil {
						return e
					}

					capool.AddCert(c)
					return nil
				}

				if fi.IsDir() {
					names, e := f.Readdirnames(0)
					if e != nil {
						fmt.Printf("Error while opening folder for CA certs %s - %v\n", OverrideCALocation, e)
					}
					for _, name := range names {
						fchild, e := os.Open(name)
						if e != nil {
							fmt.Printf("Error while loading CA cert %s - %v\n", name, e)
							fchild.Close()
							return
						}

						e = loadCA(fchild)
						fchild.Close()
						if e != nil {
							fmt.Printf("Error while loading CA cert %s - %v\n", name, e)
							return
						}
					}
				} else {
					e = loadCA(f)
					if e != nil {
						fmt.Printf("Error while loading CA cert %s - %v\n", OverrideCALocation, e)
					}
				}
			}

			c.Tls = &tls.Config{
				Certificates:       clientCert,
				InsecureSkipVerify: SkipInsecure,
				RootCAs:            capool,
			}
		} else if slc != "sbts" {
			fmt.Printf("Unknown scheme - %s is not sbts or sbtss", slc)
			return
		}

		e, reader := client.Do("tcp", u.Host, u.Path, &c)
		if e != nil {
			fmt.Printf("Error while performing transfer - %s\n", e)
			return
		}

		io.Copy(os.Stdout, reader)
		reader.Close()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	getCmd.Flags().StringVarP(&CertLocation, "cert", "t", "", "File location for the certificate the client should present")
	getCmd.Flags().StringVarP(&KeyLocation, "key", "k", "", "File location for the certificate key the client should present")
	getCmd.Flags().StringVarP(&OverrideCALocation, "ca", "a", "", "File location for an override root CA certificate")
	getCmd.Flags().BoolVarP(&SkipInsecure, "skip-insecure", "s", false, "Whether to skip authentication of the remote end")
}
