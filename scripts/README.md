# Custom Remediation Scripts

This directory contains custom remediation scripts for specific incident types.

## Script Naming Convention

Scripts should be named according to the incident type they handle:
- `SERVICE_DOWN.sh` or `SERVICE_DOWN.bat` - For service crash incidents
- `CONFIG_ERROR.sh` or `CONFIG_ERROR.bat` - For configuration issues
- `DATABASE_ERROR.sh` or `DATABASE_ERROR.bat` - For database incidents
- etc.

## Script Requirements

1. Scripts must be executable
2. Scripts should exit with code 0 on success, non-zero on failure
3. Scripts receive the incident ID as the first argument
4. Scripts can output to stdout/stderr for logging

## Example Script

```bash
#!/bin/bash
# SERVICE_DOWN.sh

INCIDENT_ID=$1
echo "Handling SERVICE_DOWN incident: $INCIDENT_ID"

# Custom remediation logic here
systemctl restart myservice

# Verify
if systemctl is-active --quiet myservice; then
    echo "Service restarted successfully"
    exit 0
else
    echo "Service restart failed"
    exit 1
fi
```

## Windows Example

```batch
@echo off
REM SERVICE_DOWN.bat

echo Handling SERVICE_DOWN incident: %1

REM Custom remediation logic here
net stop MyService
timeout /t 2
net start MyService

if %errorlevel% == 0 (
    echo Service restarted successfully
    exit /b 0
) else (
    echo Service restart failed
    exit /b 1
)
```
