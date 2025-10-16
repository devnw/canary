# API Reference

**Requirement:** {{REQ_ID}}
**Type:** API Documentation
**Created:** {{DATE}}

## Overview

High-level description of this API.

## Endpoints / Functions

### Function/Endpoint Name

**Description:** What this does

**Signature:**
```go
func FunctionName(param1 Type1, param2 Type2) (ReturnType, error)
```

**Parameters:**
- `param1` (Type1): Description
- `param2` (Type2): Description

**Returns:**
- `ReturnType`: Description of return value
- `error`: Error conditions

**Example:**
```go
result, err := FunctionName(value1, value2)
if err != nil {
    // Handle error
}
```

**Errors:**
- `ErrInvalidInput`: When input is invalid
- `ErrNotFound`: When resource not found

## Data Structures

### StructName

```go
type StructName struct {
    Field1 string
    Field2 int
}
```

**Fields:**
- `Field1`: Description
- `Field2`: Description

## Notes

- Performance considerations
- Thread safety
- Best practices
