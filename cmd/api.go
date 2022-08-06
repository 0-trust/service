/*
Copyright © 2022 Adedayo Adetoye (aka Dayo)
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"fmt"

	common "github.com/0-trust/service/pkg"
	"github.com/0-trust/service/pkg/api"
	"github.com/spf13/cobra"
)

var (
	port      int
	bindLocal bool
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run Zero Trust Modelling as an API service",
	Long:  `Run Zero Trust Modelling as an API service`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`Running %s as an API service on port %d
		
Version: %s

Author: Dr. Adedayo Adetoye (Dayo) <https://github.com/adedayo>
		`, common.AppName, port, appVersion)

		//serve API
		config := api.Config{
			AppName:    common.AppName,
			AppVersion: appVersion,
			ApiPort:    port,
			Local:      bindLocal,
		}
		api.ServeAPI(config)
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().IntVarP(&port, "port", "p", 18273, "Port on which to serve the API service")
	apiCmd.Flags().BoolVar(&bindLocal, "bind-localhost", false, "Bind the API service to localhost")

}
