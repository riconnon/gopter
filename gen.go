package gopter

import "reflect"

// Gen generator of arbitrary values
// Usually properties are checked by verifing a condition holds true for arbitrary input parameters
type Gen func(*GenParameters) *GenResult

// Sample generate a sample value
// Depending on the state of the RNG the generate might fail to provide a sample
func (g Gen) Sample() (interface{}, bool) {
	return g(DefaultGenParameters()).Retrieve()
}

func (g Gen) SuchThat(f func(interface{}) bool) Gen {
	return func(genParams *GenParameters) *GenResult {
		result := g(genParams)
		prevSieve := result.Sieve
		if prevSieve == nil {
			result.Sieve = f
		} else {
			result.Sieve = func(value interface{}) bool {
				return prevSieve(value) && f(value)
			}
		}
		return result
	}
}

func (g Gen) WithShrinker(shrinker Shrinker) Gen {
	return func(genParams *GenParameters) *GenResult {
		result := g(genParams)
		if shrinker == nil {
			result.Shrinker = NoShrinker
		} else {
			result.Shrinker = shrinker
		}
		return result
	}
}

func (g Gen) Map(f func(interface{}) interface{}) Gen {
	return func(genParams *GenParameters) *GenResult {
		result := g(genParams)
		value, ok := result.Retrieve()
		if ok {
			mapped := f(value)
			return &GenResult{
				Shrinker:   NoShrinker,
				result:     mapped,
				Labels:     result.Labels,
				ResultType: reflect.TypeOf(mapped),
			}
		}
		return &GenResult{
			Shrinker:   NoShrinker,
			result:     nil,
			Labels:     result.Labels,
			ResultType: reflect.TypeOf(nil),
		}
	}
}

func (g Gen) FlatMap(f func(interface{}) Gen) Gen {
	return func(genParams *GenParameters) *GenResult {
		result := g(genParams)
		value, ok := result.Retrieve()
		if ok {
			return f(value)(genParams)
		}
		return &GenResult{
			Shrinker:   NoShrinker,
			result:     nil,
			Labels:     result.Labels,
			ResultType: reflect.TypeOf(nil),
		}
	}
}
