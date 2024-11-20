import fs from 'fs-extra';
import path from 'path';
import { fileURLToPath } from 'url';

// Equivalent of __dirname in ES module
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Read package.json to get the target directory
const packageJsonPath = path.join(__dirname, 'package.json');
const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));

// Paths
const sourceFile = path.join(__dirname, 'dist', 'index.html');
const targetDirectory = pkg.config.targetDirectory;
const templateFile = path.join(targetDirectory, 'auth_form.html');
const testFile = path.join(__dirname, 'dist', 'index-test.html');

// Read the content of the source file
let content = fs.readFileSync(sourceFile, 'utf-8');

// Logging the initial content
console.log("Initial content length:", content.length);

// Create a version with real values for testing
let testContent = content.replace(
  /window\.formData\s*=\s*{.*?}/,
  (match) => {
    console.log("Found formData object for testing replacement:", match);
    return `window.formData = {
      chatId: "testChatID",
      messageId: "testMessageID",
      signature: "testSignature",
      authType: "testAuthType",
      csrfToken: "testCSRFToken"
    }`;
  }
);

// Create a version with placeholders for the Go project
let templateContent = content.replace(
  /window\.formData\s*=\s*{.*?}/,
  (match) => {
    console.log("Found formData object for template replacement:", match);
    return `window.formData = {
      chatId: "{{.ChatID}}",
      messageId: "{{.MessageID}}",
      signature: "{{.Signature}}",
      authType: "{{.AuthType}}",
      csrfToken: "{{.CSRFToken}}"
    }`;
  }
);

// Ensure both replacements were successful
if (!testContent.includes('testChatID')) {
  console.error("Error: Replacement for testContent failed.");
} else {
  console.log("Replacement for testContent successful.");
}

if (!templateContent.includes('{{.ChatID}}')) {
  console.error("Error: Replacement for templateContent failed.");
} else {
  console.log("Replacement for templateContent successful.");
}

// Write both files
fs.ensureDirSync(targetDirectory); // Ensure the target directory exists
fs.writeFileSync(testFile, testContent, 'utf-8');
fs.writeFileSync(templateFile, templateContent, 'utf-8');

console.log(`Template file copied to ${targetDirectory}`);
console.log(`Test file created in dist as index-test.html`);