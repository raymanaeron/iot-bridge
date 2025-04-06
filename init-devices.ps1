$baseUrl = "http://localhost:8080"

Write-Host "Starting device scan..."
Invoke-RestMethod -Method Post -Uri "$baseUrl/scan" -Headers @{ "Content-Type" = "application/json" } -Body '{"protocols": ["zigbee", "zwave"]}'

Write-Host "Waiting for scan results..."

$discovered = @()
$attempts = 0
while ($attempts -lt 10) {
    try {
        $discovered = Invoke-RestMethod -Uri "$baseUrl/scan/results"
        if ($discovered -is [System.Array] -and $discovered.Count -gt 0) { break }
        if ($discovered.id) { break }
    } catch { }
    Start-Sleep -Milliseconds 500
    $attempts++
}

if (-not $discovered -or ($discovered -isnot [System.Array] -and -not $discovered.id)) {
    Write-Host "No devices discovered. Aborting."
    exit 1
}

Write-Host "Scan results received."

$deviceMappings = @{
    "bulb1" = @{ name = "Living Room Bulb"; room = "Living Room" }
    "plug1" = @{ name = "Coffee Machine Plug"; room = "Kitchen" }
}

$deviceList = @()
if ($discovered -is [System.Array]) {
    $deviceList = $discovered
} else {
    $deviceList = @($discovered)
}

foreach ($device in $deviceList) {
    $id = $device.id
    if ($deviceMappings.ContainsKey($id)) {
        $entry = $deviceMappings[$id]
        $payload = @{
            id   = $id
            name = $entry.name
            room = $entry.room
        } | ConvertTo-Json -Compress

        Write-Host ""
        Write-Host "Registering device: $($entry.name)"
        Invoke-RestMethod -Method Post -Uri "$baseUrl/devices/from-scan" -Headers @{ "Content-Type" = "application/json" } -Body $payload
    } else {
        Write-Host "Skipping unknown device: $id"
    }
}

Write-Host ""
Write-Host "Invoking capabilities..."

Invoke-RestMethod -Method Post -Uri "$baseUrl/devices/bulb1/capabilities/power" -Headers @{ "Content-Type" = "application/json" } -Body '{"state": "on"}'
Invoke-RestMethod -Method Post -Uri "$baseUrl/devices/bulb1/capabilities/brightness" -Headers @{ "Content-Type" = "application/json" } -Body '{"level": 60}'
Invoke-RestMethod -Method Post -Uri "$baseUrl/devices/plug1/capabilities/power" -Headers @{ "Content-Type" = "application/json" } -Body '{"state": "on"}'

Write-Host ""
Write-Host "Final device list:"
Invoke-RestMethod "$baseUrl/devices" | ConvertTo-Json -Depth 5

Write-Host "`n✅ Script complete."

# 1. View /devices/bulb1
Write-Host ""
Write-Host "Retrieving device 'bulb1'..."
$response = Invoke-RestMethod -Method Get -Uri "$baseUrl/devices/bulb1"
$response | ConvertTo-Json -Depth 5

# 2. Patch /devices/bulb1
Write-Host ""
Write-Host "Updating 'bulb1' name and room..."
$patchPayload = @{
    name = "Updated Bulb"
    room = "Conference Room"
} | ConvertTo-Json -Compress

Invoke-RestMethod -Method Patch -Uri "$baseUrl/devices/bulb1" `
    -Headers @{ "Content-Type" = "application/json" } `
    -Body $patchPayload

# Confirm patch
Write-Host ""
Write-Host "Re-fetching updated 'bulb1'..."
Invoke-RestMethod -Method Get -Uri "$baseUrl/devices/bulb1" | ConvertTo-Json -Depth 5

# 3. Delete /devices/bulb1
Write-Host ""
Write-Host "Deleting device 'bulb1'..."
Invoke-RestMethod -Method Delete -Uri "$baseUrl/devices/bulb1"

# Confirm deletion
Write-Host ""
Write-Host "Final device list (should NOT include 'bulb1'):"
Invoke-RestMethod "$baseUrl/devices" | ConvertTo-Json -Depth 5
