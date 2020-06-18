package env

import (
  "fmt"
  "os"
  "reflect"
  "strconv"
)

const defaultDotEnvPath = ".env"

func Parse(v interface{}, args ...string) error {
  var varPrefix string
  var dotEnvPath string

  switch len(args) {
  case 2:
    dotEnvPath = args[1]
    fallthrough
  case 1:
    varPrefix = args[0]
  default:
    dotEnvPath = defaultDotEnvPath
  }

  if err := ParseEnvFile(v, dotEnvPath); err != nil {
    return err
  }

  if err := ParseEnv(v, varPrefix); err != nil {
    return err
  }

  return nil
}

func ParseEnv(v interface{}, args ...string) error {
  var varPrefix string

  switch len(args) {
  case 2:
    fallthrough
  case 1:
    varPrefix = args[0]
  }

  // Check that provided v is pointer to a struct
  ptr := reflect.TypeOf(v)
  typ := ptr.Elem()
  if ptr.Kind() != reflect.Ptr || typ.Kind() != reflect.Struct {
    return fmt.Errorf(
      fmt.Sprintf("expected pointer to a sturct, instead got: %T", v),
    )
  }

  val := reflect.ValueOf(v)
  if val.IsNil() {
    return fmt.Errorf("value shouldn't be nil")
  }

  parseEnv(varPrefix, typ, val.Elem())
  return nil
}

func ParseEnvFile(v interface{}, filepath string) error {
  return nil
}

func parseEnv(prefix string, typ reflect.Type, val reflect.Value) {
  // Iterate through all fields and set values
  for i := 0; i < typ.NumField(); i++ {
    field := typ.Field(i)
    fieldVal := val.Field(i)
    parseField(prefix, field, fieldVal)
  }
}

func parseEnvFile(v reflect.Value) {}

func parseField(prefix string, field reflect.StructField, val reflect.Value) {
  varName := prefix + field.Tag.Get("env")
  fmt.Println(varName)

  // Parse nested struct
  isPtr := val.Type().Kind() == reflect.Ptr && val.Type().Elem().Kind() == reflect.Struct
  isStruct := val.Type().Kind() == reflect.Struct
  if isPtr {
    if val.IsNil() {
      val.Set(reflect.New(field.Type.Elem()))
    }
    ParseEnv(val.Interface(), varName + "_")
    return
  } else if isStruct {
    ParseEnv(val.Addr().Interface(), varName + "_")
    return
  }

  required := true
  requiredVal, found := field.Tag.Lookup("required")
  fmt.Println(requiredVal)
  if found {
    required, _ = strconv.ParseBool(requiredVal)
  }

  envVal := os.Getenv(varName)
  if envVal == "" && required {
    return
  }

  if envVal == "" {
    envVal, _ = field.Tag.Lookup("default")
  }

  switch val.Type().Kind() {
  case reflect.String:
    val.SetString(envVal)
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    valInt, _ := strconv.Atoi(envVal)
    val.SetInt(int64(valInt))
  }

  return
}
