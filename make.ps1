[CmdletBinding()]
param (
    [Parameter()] [Switch] $Ci
)

if ($PSBoundParameters.count -eq 0) {
    $Ci = $True
}

if ($Ci) {
    golangci-lint run
    go test
}