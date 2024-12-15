package data

import (
	"fmt"
	"strconv"
)

// Declare a custom Runtime type, which has the underlying type int32 (the sames as our Movie struct field)
type Runtime int32

// Implement a MarshalJSON() method on the Runtime type so that it satisfies the json.Marshaler interface.
// This should return the JSON-encoded value for the movie runtime (in our case, it will return a string in the format "<runtime> mins"
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate a string containing the movie runtime in the required format
	// If your MarshalJSON() method returns a JSON string value, like our does, then you must wrap the string in double quotes before returning it.
	// Otherwise, it won't be interpreted as a JSON string, and you'll receive a runtime error similar to this.
	// We're deliberately using a value receiver for our MarshalJSON() method rather than a pointer receiver like func (r *Runtime) MarshalJSON().
	// This gives us more flexibility because it means that our custom JSON encoding will work on both Runtime values and pointers to Runtime values.
	// According to Effective Go:
	// The rule about pointers vs values for receivers is that value methods can be invoked on pointers and values, but pointer methods can only be invoked on pointers.
	jsonValue := fmt.Sprintf("%d mins", r)

	// Use the strconv.Quote() function on the string to wrap it in double quotes.
	// It needs to be surrounded by double quotes in order to be a valid *JSON string*
	quotedJSONValue := strconv.Quote(jsonValue)

	// Convert the quoted string value to a byte slice and return it.
	return []byte(quotedJSONValue), nil
}
