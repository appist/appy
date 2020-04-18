# MacOS system
.DS_Store

# Editor directories and files
.idea
.vscode
*.suo
*.ntvs*
*.njsproj
*.sln
*.sw?

# Master key files
configs/*.key
!configs/development.key
!configs/test.key

# Log files
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# 3rd party dependencies
node_modules
vendor

# Misc.
*.upx
.terraform
terraform.tfstate
terraform.tfvars
{{.projectName}}
{{.projectName}}.exe
dist
tmp
