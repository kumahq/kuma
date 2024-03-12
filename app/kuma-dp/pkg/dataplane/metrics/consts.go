package metrics

import (
	"strings"

	"github.com/prometheus/common/expfmt"
)

var FmtOpenMetrics_1_0_0 = expfmt.NewFormat(expfmt.TypeOpenMetrics)

// for some reason 0.0.1 is not taken into account in expfmt.NewFormat
var FmtOpenMetrics_0_0_1 = expfmt.Format(strings.ReplaceAll(string(FmtOpenMetrics_1_0_0), expfmt.OpenMetricsVersion_1_0_0, expfmt.OpenMetricsVersion_0_0_1))
