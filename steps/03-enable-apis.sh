#!/usr/bin/env bash
# Step 3: Enable Google APIs & Set Quota Project

step_enable_apis() {
  print_header "Step 3 of 8 — Enable Google APIs"

  # Check if token works
  if ! gcloud auth application-default print-access-token &>/dev/null; then
    print_error "No valid Google credentials — complete Step 2 first"
    return 1
  fi

  # Ensure the gcloud user account matches the active profile
  local active_profile
  active_profile=$(_get_active_profile 2>/dev/null || echo "")
  if [[ -n "$active_profile" ]]; then
    local profile_email=""
    local meta_file="$PROFILES_DIR/$active_profile/profile.json"
    if [[ -f "$meta_file" ]]; then
      profile_email=$(python3 -c "import json; d=json.load(open('$meta_file')); print(d.get('email',''))" 2>/dev/null)
    fi

    local gcloud_account
    gcloud_account=$(gcloud config get-value account 2>/dev/null)

    if [[ -n "$profile_email" && "$gcloud_account" != "$profile_email" ]]; then
      print_warn "gcloud is logged in as ${BOLD}$gcloud_account${NC}"
      print_info "Active profile '${BOLD}$active_profile${NC}' uses ${BOLD}$profile_email${NC}"
      print_blank
      print_info "Switching gcloud account to match your profile..."
      if ! gcloud auth login "$profile_email" --brief </dev/tty; then
        print_error "Failed to switch gcloud account"
        print_info "Run manually: gcloud auth login $profile_email"
        return 1
      fi
      print_success "Logged in as $profile_email"
      print_blank
    fi
  fi

  # Get project ID
  print_step "Finding your Google Cloud projects..."
  local projects
  projects=$(gcloud projects list --format="value(projectId)" 2>/dev/null)

  if [[ -z "$projects" ]]; then
    print_error "No Google Cloud projects found"
    print_info "Create one at console.cloud.google.com, then re-run"
    return 1
  fi

  local project_id
  local project_array=()
  while IFS= read -r line; do
    project_array+=("$line")
  done <<< "$projects"

  if [[ ${#project_array[@]} -eq 1 ]]; then
    project_id="${project_array[0]}"
    print_success "Found project: $project_id"
  else
    print_info "Multiple projects found. Select one:"
    project_id=$(ask_select "Choose project:" "${project_array[@]}")
  fi

  print_blank

  # Enable APIs
  local apis=(
    "drive.googleapis.com"
    "docs.googleapis.com"
    "slides.googleapis.com"
    "sheets.googleapis.com"
    "calendar-json.googleapis.com"
    "gmail.googleapis.com"
    "forms.googleapis.com"
    "tasks.googleapis.com"
    "cloudresourcemanager.googleapis.com"
  )

  print_step "Enabling Google APIs on project ${BOLD}$project_id${NC}..."
  print_info "This may take a minute on first run"
  print_blank

  if run_with_spinner "Enabling APIs..." gcloud services enable "${apis[@]}" --project="$project_id"; then
    print_success "All Google APIs enabled"
  else
    print_error "Failed to enable some APIs"
    print_info "You may need to enable billing on your GCP project"
    return 1
  fi

  # Set quota project
  print_step "Setting quota project..."

  if gcloud auth application-default set-quota-project "$project_id" 2>/dev/null; then
    print_success "Quota project set to $project_id"
  else
    print_warn "Could not set quota project automatically"
    print_info "Run manually: gcloud auth application-default set-quota-project $project_id"
  fi

  # Save project ID for later use
  mkdir -p "$HOME/.claude-vibe"
  echo "$project_id" > "$HOME/.claude-vibe/gcp-project-id"

  # Validate
  print_blank
  print_step "Validating API access..."
  local token
  token=$(gcloud auth application-default print-access-token 2>/dev/null)

  local all_ok=true
  for api_check in \
    "Drive|https://www.googleapis.com/drive/v3/about?fields=user" \
    "Gmail|https://gmail.googleapis.com/gmail/v1/users/me/profile" \
    "Calendar|https://www.googleapis.com/calendar/v3/calendars/primary" \
    "Tasks|https://tasks.googleapis.com/tasks/v1/users/@me/lists?maxResults=1"; do
    local name="${api_check%%|*}"
    local url="${api_check##*|}"
    local code
    code=$(curl -s -o /dev/null -w "%{http_code}" "$url" \
      -H "Authorization: Bearer $token" \
      -H "x-goog-user-project: $project_id" 2>/dev/null)
    case "$code" in
      200|404) print_success "$name API" ;;
      *)       print_error "$name API (HTTP $code)"; all_ok=false ;;
    esac
  done

  if $all_ok; then
    mark_step_complete "enable_apis"
  else
    print_warn "Some APIs failed — you may need to wait a few minutes for propagation"
  fi
}
