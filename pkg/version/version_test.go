package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {

	ver := Core()
	assert.Equal(t, coreVersion, ver)

	sh := Short()
	assert.Equal(t, coreVersion, sh)

	commit = "sha"
	f := Full()
	assert.Equal(t, "v", string(f[0]))
}
