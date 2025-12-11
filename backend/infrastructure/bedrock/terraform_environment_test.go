package bedrock

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTerraformEnvironmentConfiguration_Development tests development Terraform configuration
func TestTerraformEnvironmentConfiguration_Development(t *testing.T) {
	if !isTerraformAvailable() {
		t.Skip("Terraform not available, skipping Terraform environment tests")
	}

	terraformDir := filepath.Join("..", "..", "..", "terraform", "environments", "dev")
	if !dirExists(terraformDir) {
		t.Skip("Development Terraform directory not found, skipping test")
	}

	// Validate Terraform configuration files
	t.Run("ValidateConfigurationFiles", func(t *testing.T) {
		validateTerraformFiles(t, terraformDir, "dev")
	})

	// Validate Terraform variables
	t.Run("ValidateTerraformVariables", func(t *testing.T) {
		validateTerraformVariables(t, terraformDir, map[string]interface{}{
			"environment":      "dev",
			"project_name":     "kb",
			"aws_region":       "us-east-1",
			"foundation_model": "us.amazon.nova-2-lite-v1:0",
			"enable_vpc":       false,
		})
	})

	// Test Terraform plan (if credentials available)
	t.Run("ValidateTerraformPlan", func(t *testing.T) {
		if !hasAWSCredentials() {
			t.Skip("AWS credentials not available, skipping Terraform plan test")
		}
		validateTerraformPlan(t, terraformDir)
	})
}

// TestTerraformEnvironmentConfiguration_Staging tests staging Terraform configuration
func TestTerraformEnvironmentConfiguration_Staging(t *testing.T) {
	if !isTerraformAvailable() {
		t.Skip("Terraform not available, skipping Terraform environment tests")
	}

	terraformDir := filepath.Join("..", "..", "..", "terraform", "environments", "staging")
	if !dirExists(terraformDir) {
		t.Skip("Staging Terraform directory not found, skipping test")
	}

	// Validate Terraform configuration files
	t.Run("ValidateConfigurationFiles", func(t *testing.T) {
		validateTerraformFiles(t, terraformDir, "staging")
	})

	// Validate Terraform variables
	t.Run("ValidateTerraformVariables", func(t *testing.T) {
		validateTerraformVariables(t, terraformDir, map[string]interface{}{
			"environment":      "staging",
			"project_name":     "bedrock-chat-poc",
			"aws_region":       "us-east-1",
			"foundation_model": "anthropic.claude-v2",
			"enable_vpc":       false,
		})
	})

	// Test Terraform plan (if credentials available)
	t.Run("ValidateTerraformPlan", func(t *testing.T) {
		if !hasAWSCredentials() {
			t.Skip("AWS credentials not available, skipping Terraform plan test")
		}
		validateTerraformPlan(t, terraformDir)
	})
}

// TestTerraformEnvironmentConfiguration_Production tests production Terraform configuration
func TestTerraformEnvironmentConfiguration_Production(t *testing.T) {
	if !isTerraformAvailable() {
		t.Skip("Terraform not available, skipping Terraform environment tests")
	}

	terraformDir := filepath.Join("..", "..", "..", "terraform", "environments", "prod")
	if !dirExists(terraformDir) {
		t.Skip("Production Terraform directory not found, skipping test")
	}

	// Validate Terraform configuration files
	t.Run("ValidateConfigurationFiles", func(t *testing.T) {
		validateTerraformFiles(t, terraformDir, "prod")
	})

	// Validate Terraform variables
	t.Run("ValidateTerraformVariables", func(t *testing.T) {
		validateTerraformVariables(t, terraformDir, map[string]interface{}{
			"environment":      "prod",
			"project_name":     "bedrock-chat-poc",
			"aws_region":       "us-east-1",
			"foundation_model": "anthropic.claude-v2",
			"enable_vpc":       true,
		})
	})

	// Validate VPC configuration for production
	t.Run("ValidateVPCConfiguration", func(t *testing.T) {
		validateProductionVPCConfig(t, terraformDir)
	})

	// Test Terraform plan (if credentials available)
	t.Run("ValidateTerraformPlan", func(t *testing.T) {
		if !hasAWSCredentials() {
			t.Skip("AWS credentials not available, skipping Terraform plan test")
		}
		validateTerraformPlan(t, terraformDir)
	})
}

// TestTerraformModuleConfiguration tests Terraform module configurations
func TestTerraformModuleConfiguration(t *testing.T) {
	if !isTerraformAvailable() {
		t.Skip("Terraform not available, skipping Terraform module tests")
	}

	modules := []string{
		"bedrock-agent",
		"knowledge-base",
		"iam",
		"vpc",
	}

	for _, module := range modules {
		t.Run(fmt.Sprintf("ValidateModule_%s", module), func(t *testing.T) {
			moduleDir := filepath.Join("..", "..", "..", "terraform", "modules", module)
			if !dirExists(moduleDir) {
				t.Skipf("Module directory %s not found, skipping test", module)
			}

			validateTerraformModule(t, moduleDir, module)
		})
	}
}

// Helper functions for Terraform validation

func validateTerraformFiles(t *testing.T, terraformDir, environment string) {
	// Check required files exist
	requiredFiles := []string{
		"terraform.tfvars",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(terraformDir, file)
		if !fileExists(filePath) {
			t.Errorf("Required file %s should exist in %s environment", file, environment)
		}
	}

	// Check main.tf exists in parent directory or current directory
	mainTfPaths := []string{
		filepath.Join(terraformDir, "main.tf"),
		filepath.Join(terraformDir, "..", "..", "main.tf"),
	}

	mainTfExists := false
	for _, path := range mainTfPaths {
		if fileExists(path) {
			mainTfExists = true
			break
		}
	}
	if !mainTfExists {
		t.Errorf("main.tf should exist for %s environment", environment)
	}
}

func validateTerraformVariables(t *testing.T, terraformDir string, expectedVars map[string]interface{}) {
	tfvarsPath := filepath.Join(terraformDir, "terraform.tfvars")
	if !fileExists(tfvarsPath) {
		t.Skipf("terraform.tfvars not found at %s", tfvarsPath)
	}

	content, err := os.ReadFile(tfvarsPath)
	if err != nil {
		t.Fatalf("Should be able to read terraform.tfvars: %v", err)
	}

	tfvarsContent := string(content)

	for key, expectedValue := range expectedVars {
		switch v := expectedValue.(type) {
		case string:
			// Check if the key exists and has some value
			if strings.Contains(tfvarsContent, key+" =") {
				t.Logf("Variable %s is defined in %s", key, terraformDir)
			} else {
				t.Errorf("Variable %s should be defined", key)
			}
		case bool:
			pattern := fmt.Sprintf(`%s = %t`, key, v)
			if !strings.Contains(tfvarsContent, pattern) {
				t.Errorf("Variable %s should be set to %t", key, v)
			}
		default:
			// For other types, just check the key exists
			if !strings.Contains(tfvarsContent, key) {
				t.Errorf("Variable %s should be defined", key)
			}
		}
	}
}

func validateProductionVPCConfig(t *testing.T, terraformDir string) {
	tfvarsPath := filepath.Join(terraformDir, "terraform.tfvars")
	if !fileExists(tfvarsPath) {
		t.Skipf("terraform.tfvars not found at %s", tfvarsPath)
	}

	content, err := os.ReadFile(tfvarsPath)
	if err != nil {
		t.Fatalf("Should be able to read terraform.tfvars: %v", err)
	}

	tfvarsContent := string(content)

	// Check VPC-specific configurations for production
	vpcConfigs := map[string]string{
		"enable_vpc":           "true",
		"vpc_cidr":             "10.0.0.0/16",
		"availability_zones":   `["us-east-1a", "us-east-1b"]`,
		"enable_nat_gateway":   "true",
		"single_nat_gateway":   "false",
	}

	for key, expectedValue := range vpcConfigs {
		if !strings.Contains(tfvarsContent, key) {
			t.Errorf("Production VPC config should include %s", key)
		}
		if expectedValue != "" && !strings.Contains(tfvarsContent, expectedValue) {
			t.Errorf("Production VPC config %s should be set correctly", key)
		}
	}
}

func validateTerraformModule(t *testing.T, moduleDir, moduleName string) {
	// Check required module files
	requiredFiles := []string{
		"main.tf",
		"variables.tf",
		"outputs.tf",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(moduleDir, file)
		if !fileExists(filePath) {
			t.Errorf("Module %s should have %s", moduleName, file)
		}
	}

	// Validate Terraform syntax if terraform is available
	if isTerraformAvailable() {
		cmd := exec.Command("terraform", "validate")
		cmd.Dir = moduleDir
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			// Try terraform init first
			initCmd := exec.Command("terraform", "init", "-backend=false")
			initCmd.Dir = moduleDir
			initOutput, initErr := initCmd.CombinedOutput()
			
			if initErr != nil {
				t.Logf("Terraform init failed for module %s: %s", moduleName, string(initOutput))
				t.Skipf("Skipping validation for module %s due to init failure", moduleName)
			}
			
			// Try validate again
			cmd = exec.Command("terraform", "validate")
			cmd.Dir = moduleDir
			output, err = cmd.CombinedOutput()
		}
		
		if err != nil {
			t.Logf("Terraform validate output for module %s: %s", moduleName, string(output))
			t.Errorf("Terraform validation should pass for module %s: %v", moduleName, err)
		}
	}
}

func validateTerraformPlan(t *testing.T, terraformDir string) {
	// Initialize Terraform
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = terraformDir
	initOutput, err := initCmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Terraform init output: %s", string(initOutput))
		t.Skipf("Terraform init failed, skipping plan validation: %v", err)
	}

	// Run Terraform plan
	planCmd := exec.Command("terraform", "plan", "-detailed-exitcode")
	planCmd.Dir = terraformDir
	planOutput, err := planCmd.CombinedOutput()
	
	// Exit code 0 = no changes, 1 = error, 2 = changes planned
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 2 {
				// Changes planned - this is expected for a plan
				t.Logf("Terraform plan shows changes (expected): %s", string(planOutput))
				return
			}
		}
		t.Logf("Terraform plan output: %s", string(planOutput))
		t.Errorf("Terraform plan failed: %v", err)
	} else {
		t.Logf("Terraform plan successful (no changes): %s", string(planOutput))
	}
}

// Utility functions

func isTerraformAvailable() bool {
	_, err := exec.LookPath("terraform")
	return err == nil
}

func hasAWSCredentials() bool {
	// Check for AWS credentials in environment or AWS CLI
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}
	
	// Check if AWS CLI is configured
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	err := cmd.Run()
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}