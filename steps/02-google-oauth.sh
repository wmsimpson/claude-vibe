#!/usr/bin/env bash
# Step 2: Google OAuth Setup

step_google_oauth() {
  print_header "Step 2 of 8 — Google OAuth Setup"

  local cred_dir="$HOME/.config/gcloud/credentials"
  local cred_file="$cred_dir/claude-google-auth.json"

  # Check if credentials already exist
  if [[ -f "$cred_file" ]]; then
    print_success "OAuth credentials found at $cred_file"
    if ask_yes_no "Skip OAuth setup and use existing credentials?" "y"; then
      _run_google_auth "$cred_file"
      mark_step_complete "google_oauth"
      return 0
    fi
  fi

  # Check if ADC already configured
  if gcloud auth application-default print-access-token &>/dev/null; then
    print_success "Application Default Credentials already configured"
    if ask_yes_no "Skip OAuth setup?" "y"; then
      mark_step_complete "google_oauth"
      return 0
    fi
  fi

  print_blank
  print_info "Claude Vibe needs a Google OAuth client to access Google Workspace."
  print_info "You'll create a free OAuth client in Google Cloud Console."
  print_blank

  echo -e "  ${BOLD}Follow these steps in your browser:${NC}"
  print_blank
  echo -e "  ${CYAN}1.${NC} Go to ${BOLD}console.cloud.google.com${NC}"
  echo -e "  ${CYAN}2.${NC} Create a project (e.g. ${BOLD}claude-code-local${NC}) or select an existing one"
  echo -e "  ${CYAN}3.${NC} Go to ${BOLD}APIs & Services → Credentials${NC}"
  echo -e "  ${CYAN}4.${NC} Click ${BOLD}Create Credentials → OAuth client ID${NC}"
  echo -e "  ${CYAN}5.${NC} Select ${BOLD}Desktop app${NC}, name it ${BOLD}claude-google-auth${NC}"
  echo -e "  ${CYAN}6.${NC} Add redirect URIs: ${BOLD}http://localhost:8086/${NC} and ${BOLD}http://localhost:8088/${NC}"
  echo -e "  ${CYAN}7.${NC} Download the JSON file to ${BOLD}~/Downloads/${NC}"
  print_blank

  # Wait for user to complete browser steps
  if ! ask_yes_no "Have you completed the steps above and downloaded the JSON?" "n"; then
    print_warn "Skipping Google OAuth — you can run this step again later"
    return 0
  fi

  # Find the downloaded file
  local download_path
  if [[ -f "$HOME/Downloads/claude-google-auth.json" ]]; then
    download_path="$HOME/Downloads/claude-google-auth.json"
  elif [[ -f "$HOME/Downloads/client_secret_*.json" ]]; then
    download_path=$(ls -t "$HOME/Downloads/client_secret_"*.json 2>/dev/null | head -1)
  else
    download_path=$(ask_input "Path to downloaded OAuth JSON" "$HOME/Downloads/claude-google-auth.json")
  fi

  if [[ ! -f "$download_path" ]]; then
    print_error "File not found: $download_path"
    print_info "Move your OAuth JSON to $cred_file manually, then re-run"
    return 1
  fi

  # Move and secure
  mkdir -p "$cred_dir"
  cp "$download_path" "$cred_file"
  chmod 600 "$cred_file"
  chmod 700 "$cred_dir"
  print_success "Credentials secured at $cred_file"

  # Run auth
  _run_google_auth "$cred_file"
  mark_step_complete "google_oauth"
}

_run_google_auth() {
  local cred_file="$1"
  local scopes="https://www.googleapis.com/auth/drive"
  scopes+=",https://www.googleapis.com/auth/documents"
  scopes+=",https://www.googleapis.com/auth/presentations"
  scopes+=",https://www.googleapis.com/auth/spreadsheets"
  scopes+=",https://www.googleapis.com/auth/calendar"
  scopes+=",https://www.googleapis.com/auth/gmail.modify"
  scopes+=",https://www.googleapis.com/auth/forms.body"
  scopes+=",https://www.googleapis.com/auth/forms.responses.readonly"
  scopes+=",https://www.googleapis.com/auth/tasks"
  scopes+=",https://www.googleapis.com/auth/cloud-platform"

  print_blank
  print_step "Launching Google OAuth in your browser..."
  print_info "Complete the sign-in and click ${BOLD}Allow${NC}"
  print_blank

  if gcloud auth application-default login \
    --client-id-file="$cred_file" \
    --scopes="$scopes" </dev/tty; then
    print_success "Google authentication complete"
  else
    print_error "Authentication failed — re-run setup to try again"
    print_info "If you see 'This app is blocked', check your redirect URIs"
    return 1
  fi
}
