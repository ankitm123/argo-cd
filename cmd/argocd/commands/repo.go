package commands

import (
	stderrors "errors"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/argoproj/argo-cd/v3/cmd/argocd/commands/headless"
	"github.com/argoproj/argo-cd/v3/cmd/argocd/commands/utils"
	cmdutil "github.com/argoproj/argo-cd/v3/cmd/util"
	argocdclient "github.com/argoproj/argo-cd/v3/pkg/apiclient"
	repositorypkg "github.com/argoproj/argo-cd/v3/pkg/apiclient/repository"
	appsv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v3/util/cli"
	"github.com/argoproj/argo-cd/v3/util/errors"
	"github.com/argoproj/argo-cd/v3/util/git"
	utilio "github.com/argoproj/argo-cd/v3/util/io"
)

// NewRepoCommand returns a new instance of an `argocd repo` command
func NewRepoCommand(clientOpts *argocdclient.ClientOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "repo",
		Short: "Manage repository connection parameters",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
			os.Exit(1)
		},
		Example: `
# Add git repository connection parameters
argocd repo add git@git.example.com:repos/repo

# Get a Configured Repository by URL
argocd repo get https://github.com/yourusername/your-repo.git

# List Configured Repositories
argocd repo list

# Remove Configured Repositories
argocd repo rm https://github.com/yourusername/your-repo.git
`,
	}

	command.AddCommand(NewRepoAddCommand(clientOpts))
	command.AddCommand(NewRepoGetCommand(clientOpts))
	command.AddCommand(NewRepoListCommand(clientOpts))
	command.AddCommand(NewRepoRemoveCommand(clientOpts))
	return command
}

// NewRepoAddCommand returns a new instance of an `argocd repo add` command
func NewRepoAddCommand(clientOpts *argocdclient.ClientOptions) *cobra.Command {
	var repoOpts cmdutil.RepoOptions

	// For better readability and easier formatting
	repoAddExamples := `  # Add a Git repository via SSH using a private key for authentication, ignoring the server's host key:
  argocd repo add git@git.example.com:repos/repo --insecure-ignore-host-key --ssh-private-key-path ~/id_rsa

  # Add a Git repository via SSH on a non-default port - need to use ssh:// style URLs here
  argocd repo add ssh://git@git.example.com:2222/repos/repo --ssh-private-key-path ~/id_rsa

  # Add a Git repository via SSH using socks5 proxy with no proxy credentials
  argocd repo add ssh://git@github.com/argoproj/argocd-example-apps --ssh-private-key-path ~/id_rsa --proxy socks5://your.proxy.server.ip:1080

  # Add a Git repository via SSH using socks5 proxy with proxy credentials
  argocd repo add ssh://git@github.com/argoproj/argocd-example-apps --ssh-private-key-path ~/id_rsa --proxy socks5://username:password@your.proxy.server.ip:1080

  # Add a private Git repository via HTTPS using username/password and TLS client certificates:
  argocd repo add https://git.example.com/repos/repo --username git --password secret --tls-client-cert-path ~/mycert.crt --tls-client-cert-key-path ~/mycert.key

  # Add a private Git repository via HTTPS using username/password without verifying the server's TLS certificate
  argocd repo add https://git.example.com/repos/repo --username git --password secret --insecure-skip-server-verification

  # Add a public Helm repository named 'stable' via HTTPS
  argocd repo add https://charts.helm.sh/stable --type helm --name stable  

  # Add a private Helm repository named 'stable' via HTTPS
  argocd repo add https://charts.helm.sh/stable --type helm --name stable --username test --password test

  # Add a private Helm OCI-based repository named 'stable' via HTTPS
  argocd repo add helm-oci-registry.cn-zhangjiakou.cr.aliyuncs.com --type helm --name stable --enable-oci --username test --password test
  
  # Add a private HTTPS OCI repository named 'stable'
  argocd repo add oci://helm-oci-registry.cn-zhangjiakou.cr.aliyuncs.com --type oci --name stable --username test --password test
  
  # Add a private OCI repository named 'stable' without verifying the server's TLS certificate
  argocd repo add oci://helm-oci-registry.cn-zhangjiakou.cr.aliyuncs.com --type oci --name stable --username test --password test --insecure-skip-server-verification
  
  # Add a private HTTP OCI repository named 'stable'
  argocd repo add oci://helm-oci-registry.cn-zhangjiakou.cr.aliyuncs.com --type oci --name stable --username test --password test --insecure-oci-force-http

  # Add a private Git repository on GitHub.com via GitHub App
  argocd repo add https://git.example.com/repos/repo --github-app-id 1 --github-app-installation-id 2 --github-app-private-key-path test.private-key.pem

  # Add a private Git repository on GitHub Enterprise via GitHub App
  argocd repo add https://ghe.example.com/repos/repo --github-app-id 1 --github-app-installation-id 2 --github-app-private-key-path test.private-key.pem --github-app-enterprise-base-url https://ghe.example.com/api/v3

  # Add a private Git repository on Google Cloud Sources via GCP service account credentials
  argocd repo add https://source.developers.google.com/p/my-google-cloud-project/r/my-repo --gcp-service-account-key-path service-account-key.json
`

	command := &cobra.Command{
		Use:     "add REPOURL",
		Short:   "Add git, oci or helm repository connection parameters",
		Example: repoAddExamples,
		Run: func(c *cobra.Command, args []string) {
			ctx := c.Context()

			if len(args) != 1 {
				c.HelpFunc()(c, args)
				os.Exit(1)
			}

			// Repository URL
			repoOpts.Repo.Repo = args[0]

			// Specifying ssh-private-key-path is only valid for SSH repositories
			if repoOpts.SshPrivateKeyPath != "" {
				if ok, _ := git.IsSSHURL(repoOpts.Repo.Repo); ok {
					keyData, err := os.ReadFile(repoOpts.SshPrivateKeyPath)
					if err != nil {
						log.Fatal(err)
					}
					repoOpts.Repo.SSHPrivateKey = string(keyData)
				} else {
					errors.Fatal(errors.ErrorGeneric, "--ssh-private-key-path is only supported for SSH repositories.")
				}
			}

			// tls-client-cert-path and tls-client-cert-key-key-path must always be
			// specified together
			if (repoOpts.TlsClientCertPath != "" && repoOpts.TlsClientCertKeyPath == "") || (repoOpts.TlsClientCertPath == "" && repoOpts.TlsClientCertKeyPath != "") {
				err := stderrors.New("--tls-client-cert-path and --tls-client-cert-key-path must be specified together")
				errors.CheckError(err)
			}

			// Specifying tls-client-cert-path is only valid for HTTPS repositories
			if repoOpts.TlsClientCertPath != "" {
				if git.IsHTTPSURL(repoOpts.Repo.Repo) {
					tlsCertData, err := os.ReadFile(repoOpts.TlsClientCertPath)
					errors.CheckError(err)
					tlsCertKey, err := os.ReadFile(repoOpts.TlsClientCertKeyPath)
					errors.CheckError(err)
					repoOpts.Repo.TLSClientCertData = string(tlsCertData)
					repoOpts.Repo.TLSClientCertKey = string(tlsCertKey)
				} else {
					err := stderrors.New("--tls-client-cert-path is only supported for HTTPS repositories")
					errors.CheckError(err)
				}
			}

			// Specifying github-app-private-key-path is only valid for HTTPS repositories
			if repoOpts.GithubAppPrivateKeyPath != "" {
				if git.IsHTTPSURL(repoOpts.Repo.Repo) {
					githubAppPrivateKey, err := os.ReadFile(repoOpts.GithubAppPrivateKeyPath)
					errors.CheckError(err)
					repoOpts.Repo.GithubAppPrivateKey = string(githubAppPrivateKey)
				} else {
					err := stderrors.New("--github-app-private-key-path is only supported for HTTPS repositories")
					errors.CheckError(err)
				}
			}

			if repoOpts.GCPServiceAccountKeyPath != "" {
				if git.IsHTTPSURL(repoOpts.Repo.Repo) {
					gcpServiceAccountKey, err := os.ReadFile(repoOpts.GCPServiceAccountKeyPath)
					errors.CheckError(err)
					repoOpts.Repo.GCPServiceAccountKey = string(gcpServiceAccountKey)
				} else {
					err := stderrors.New("--gcp-service-account-key-path is only supported for HTTPS repositories")
					errors.CheckError(err)
				}
			}

			// Set repository connection properties only when creating repository, not
			// when creating repository credentials.
			// InsecureIgnoreHostKey is deprecated and only here for backwards compat
			repoOpts.Repo.InsecureIgnoreHostKey = repoOpts.InsecureIgnoreHostKey
			repoOpts.Repo.Insecure = repoOpts.InsecureSkipServerVerification
			repoOpts.Repo.EnableLFS = repoOpts.EnableLfs
			repoOpts.Repo.EnableOCI = repoOpts.EnableOci
			repoOpts.Repo.GithubAppId = repoOpts.GithubAppId
			repoOpts.Repo.GithubAppInstallationId = repoOpts.GithubAppInstallationId
			repoOpts.Repo.GitHubAppEnterpriseBaseURL = repoOpts.GitHubAppEnterpriseBaseURL
			repoOpts.Repo.Proxy = repoOpts.Proxy
			repoOpts.Repo.NoProxy = repoOpts.NoProxy
			repoOpts.Repo.ForceHttpBasicAuth = repoOpts.ForceHttpBasicAuth
			repoOpts.Repo.UseAzureWorkloadIdentity = repoOpts.UseAzureWorkloadIdentity

			if repoOpts.Repo.Type == "helm" && repoOpts.Repo.Name == "" {
				errors.Fatal(errors.ErrorGeneric, "Must specify --name for repos of type 'helm'")
			}

			if repoOpts.Repo.Type == "oci" && repoOpts.InsecureOCIForceHTTP {
				repoOpts.Repo.InsecureOCIForceHttp = repoOpts.InsecureOCIForceHTTP
			}

			conn, repoIf := headless.NewClientOrDie(clientOpts, c).NewRepoClientOrDie()
			defer utilio.Close(conn)

			// If the user set a username, but didn't supply password via --password,
			// then we prompt for it
			if repoOpts.Repo.Username != "" && repoOpts.Repo.Password == "" {
				repoOpts.Repo.Password = cli.PromptPassword(repoOpts.Repo.Password)
			}

			err := cmdutil.ValidateBearerTokenAndPasswordCombo(repoOpts.Repo.BearerToken, repoOpts.Repo.Password)
			errors.CheckError(err)
			err = cmdutil.ValidateBearerTokenForGitOnly(repoOpts.Repo.BearerToken, repoOpts.Repo.Type)
			errors.CheckError(err)
			err = cmdutil.ValidateBearerTokenForHTTPSRepoOnly(repoOpts.Repo.BearerToken, git.IsHTTPSURL(repoOpts.Repo.Repo))
			errors.CheckError(err)

			// We let the server check access to the repository before adding it. If
			// it is a private repo, but we cannot access with the credentials
			// that were supplied, we bail out.
			//
			// Skip validation if we are just adding credentials template, chances
			// are high that we do not have the given URL pointing to a valid Git
			// repo anyway.
			repoAccessReq := repositorypkg.RepoAccessQuery{
				Repo:                       repoOpts.Repo.Repo,
				Type:                       repoOpts.Repo.Type,
				Name:                       repoOpts.Repo.Name,
				Username:                   repoOpts.Repo.Username,
				Password:                   repoOpts.Repo.Password,
				BearerToken:                repoOpts.Repo.BearerToken,
				SshPrivateKey:              repoOpts.Repo.SSHPrivateKey,
				TlsClientCertData:          repoOpts.Repo.TLSClientCertData,
				TlsClientCertKey:           repoOpts.Repo.TLSClientCertKey,
				Insecure:                   repoOpts.Repo.IsInsecure(),
				EnableOci:                  repoOpts.Repo.EnableOCI,
				GithubAppPrivateKey:        repoOpts.Repo.GithubAppPrivateKey,
				GithubAppID:                repoOpts.Repo.GithubAppId,
				GithubAppInstallationID:    repoOpts.Repo.GithubAppInstallationId,
				GithubAppEnterpriseBaseUrl: repoOpts.Repo.GitHubAppEnterpriseBaseURL,
				Proxy:                      repoOpts.Proxy,
				Project:                    repoOpts.Repo.Project,
				GcpServiceAccountKey:       repoOpts.Repo.GCPServiceAccountKey,
				ForceHttpBasicAuth:         repoOpts.Repo.ForceHttpBasicAuth,
				UseAzureWorkloadIdentity:   repoOpts.Repo.UseAzureWorkloadIdentity,
				InsecureOciForceHttp:       repoOpts.Repo.InsecureOCIForceHttp,
			}
			_, err = repoIf.ValidateAccess(ctx, &repoAccessReq)
			errors.CheckError(err)

			repoCreateReq := repositorypkg.RepoCreateRequest{
				Repo:   &repoOpts.Repo,
				Upsert: repoOpts.Upsert,
			}

			createdRepo, err := repoIf.CreateRepository(ctx, &repoCreateReq)
			errors.CheckError(err)
			fmt.Printf("Repository '%s' added\n", createdRepo.Repo)
		},
	}
	command.Flags().BoolVar(&repoOpts.Upsert, "upsert", false, "Override an existing repository with the same name even if the spec differs")
	cmdutil.AddRepoFlags(command, &repoOpts)
	return command
}

// NewRepoRemoveCommand returns a new instance of an `argocd repo rm` command
func NewRepoRemoveCommand(clientOpts *argocdclient.ClientOptions) *cobra.Command {
	var project string
	command := &cobra.Command{
		Use:   "rm REPO ...",
		Short: "Remove configured repositories",
		Example: `
  # Remove a single repository
  argocd repo rm https://github.com/yourusername/your-repo.git

  # Remove multiple repositories
  argocd repo rm https://github.com/yourusername/your-repo.git https://git.example.com/repo2.git

  # Remove repositories for a specific project
  argocd repo rm https://github.com/yourusername/your-repo.git --project myproject

  # Remove repository using SSH URL
  argocd repo rm git@github.com:yourusername/your-repo.git
`,
		Run: func(c *cobra.Command, args []string) {
			ctx := c.Context()

			if len(args) == 0 {
				c.HelpFunc()(c, args)
				os.Exit(1)
			}
			conn, repoIf := headless.NewClientOrDie(clientOpts, c).NewRepoClientOrDie()
			defer utilio.Close(conn)

			promptUtil := utils.NewPrompt(clientOpts.PromptsEnabled)
			for _, repoURL := range args {
				canDelete := promptUtil.Confirm(fmt.Sprintf("Are you sure you want to delete repository '%s'? [y/n]", repoURL))
				if canDelete {
					_, err := repoIf.DeleteRepository(ctx, &repositorypkg.RepoQuery{Repo: repoURL, AppProject: project})
					errors.CheckError(err)
					fmt.Printf("Repository '%s' removed\n", repoURL)
				} else {
					fmt.Printf("The command to delete '%s' was cancelled.\n", repoURL)
				}
			}
		},
	}
	command.Flags().StringVar(&project, "project", "", "project of the repository")
	return command
}

// Print table of repo info
func printRepoTable(repos appsv1.Repositories) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "TYPE\tNAME\tREPO\tINSECURE\tOCI\tLFS\tCREDS\tSTATUS\tMESSAGE\tPROJECT\n")
	for _, r := range repos {
		var hasCreds string
		if r.InheritedCreds {
			hasCreds = "inherited"
		} else {
			hasCreds = strconv.FormatBool(r.HasCredentials())
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%v\t%v\t%s\t%s\t%s\t%s\n", r.Type, r.Name, r.Repo, r.IsInsecure(), r.EnableOCI, r.EnableLFS, hasCreds, r.ConnectionState.Status, r.ConnectionState.Message, r.Project)
	}
	_ = w.Flush()
}

// Print list of repo urls or url patterns for repository credentials
func printRepoUrls(repos appsv1.Repositories) {
	for _, r := range repos {
		fmt.Println(r.Repo)
	}
}

// NewRepoListCommand returns a new instance of an `argocd repo list` command
func NewRepoListCommand(clientOpts *argocdclient.ClientOptions) *cobra.Command {
	var (
		output  string
		refresh string
	)
	command := &cobra.Command{
		Use:   "list",
		Short: "List configured repositories",
		Example: `
  # List all repositories
  argocd repo list

  # List repositories in wide format
  argocd repo list -o wide

  # List repositories in YAML format
  argocd repo list -o yaml

  # List repositories in JSON format
  argocd repo list -o json

  # List urls of repositories
  argocd repo list -o url

  # Force refresh of cached repository connection status
  argocd repo list --refresh hard
`,
		Run: func(c *cobra.Command, _ []string) {
			ctx := c.Context()

			conn, repoIf := headless.NewClientOrDie(clientOpts, c).NewRepoClientOrDie()
			defer utilio.Close(conn)
			forceRefresh := false
			switch refresh {
			case "":
			case "hard":
				forceRefresh = true
			default:
				err := stderrors.New("--refresh must be one of: 'hard'")
				errors.CheckError(err)
			}
			repos, err := repoIf.ListRepositories(ctx, &repositorypkg.RepoQuery{ForceRefresh: forceRefresh})
			errors.CheckError(err)
			switch output {
			case "yaml", "json":
				err := PrintResourceList(repos.Items, output, false)
				errors.CheckError(err)
			case "url":
				printRepoUrls(repos.Items)
				// wide is the default
			case "wide", "":
				printRepoTable(repos.Items)
			default:
				errors.CheckError(fmt.Errorf("unknown output format: %s", output))
			}
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "wide", "Output format. One of: json|yaml|wide|url")
	command.Flags().StringVar(&refresh, "refresh", "", "Force a cache refresh on connection status , must be one of: 'hard'")
	return command
}

// NewRepoGetCommand returns a new instance of an `argocd repo get` command
func NewRepoGetCommand(clientOpts *argocdclient.ClientOptions) *cobra.Command {
	var (
		output  string
		refresh string
		project string
	)

	// For better readability and easier formatting
	repoGetExamples := `
  # Get Git or Helm repository details in wide format (default, '-o wide')
  argocd repo get https://git.example.com/repos/repo

  # Get repository details in YAML format
  argocd repo get https://git.example.com/repos/repo -o yaml

  # Get repository details in JSON format
  argocd repo get https://git.example.com/repos/repo -o json

  # Get repository URL
  argocd repo get https://git.example.com/repos/repo -o url
`

	command := &cobra.Command{
		Use:     "get REPO",
		Short:   "Get a configured repository by URL",
		Example: repoGetExamples,
		Run: func(c *cobra.Command, args []string) {
			ctx := c.Context()

			if len(args) != 1 {
				c.HelpFunc()(c, args)
				os.Exit(1)
			}

			// Repository URL
			repoURL := args[0]
			conn, repoIf := headless.NewClientOrDie(clientOpts, c).NewRepoClientOrDie()
			defer utilio.Close(conn)
			forceRefresh := false
			switch refresh {
			case "":
			case "hard":
				forceRefresh = true
			default:
				err := stderrors.New("--refresh must be one of: 'hard'")
				errors.CheckError(err)
			}
			repo, err := repoIf.Get(ctx, &repositorypkg.RepoQuery{Repo: repoURL, ForceRefresh: forceRefresh, AppProject: project})
			errors.CheckError(err)
			switch output {
			case "yaml", "json":
				err := PrintResource(repo, output)
				errors.CheckError(err)
			case "url":
				fmt.Println(repo.Repo)
				// wide is the default
			case "wide", "":
				printRepoTable(appsv1.Repositories{repo})
			default:
				errors.CheckError(fmt.Errorf("unknown output format: %s", output))
			}
		},
	}

	command.Flags().StringVar(&project, "project", "", "project of the repository")
	command.Flags().StringVarP(&output, "output", "o", "wide", "Output format. One of: json|yaml|wide|url")
	command.Flags().StringVar(&refresh, "refresh", "", "Force a cache refresh on connection status , must be one of: 'hard'")
	return command
}
