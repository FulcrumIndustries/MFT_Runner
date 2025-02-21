import React, { useState, useMemo } from "react";
import {
  Box,
  Button,
  Card,
  CardContent,
  Grid,
  LinearProgress,
  Typography,
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  CircularProgress,
} from "@mui/material";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  LineChart,
  Line,
  AreaChart,
  Area,
} from "recharts";
import {
  Upload as UploadIcon,
  Refresh as RefreshIcon,
  BrightnessLow as BrightnessLowIcon,
  Brightness4 as Brightness4Icon,
  Brightness7 as Brightness7Icon,
  Upload as UploadIconMUI,
  Download as DownloadIcon,
} from "@mui/icons-material";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import { useDropzone } from "react-dropzone";
import {
  alpha,
  useTheme,
  createTheme,
  ThemeProvider,
} from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";

const formatDuration = (ms) => `${(ms / 1000).toFixed(2)}s`;
const formatSize = (size) => `${size} KB`;

const formatDurationCompact = (ms) => {
  const seconds = ms / 1000;
  if (seconds > 60) {
    return `${Math.floor(seconds / 60)}m ${Math.round(seconds % 60)}s`;
  }
  return `${seconds.toFixed(1)}s`;
};

const generateHistogramData = (latencies, binCount = 20) => {
  if (!latencies || latencies.length === 0) return [];

  const min = Math.min(...latencies);
  const max = Math.max(...latencies);
  const binSize = (max - min) / binCount;

  return Array.from({ length: binCount }, (_, i) => {
    const binStart = min + i * binSize;
    const binEnd = binStart + binSize;
    return {
      range: `${Math.round(binStart)}-${Math.round(binEnd)}ms`,
      count: latencies.filter((l) => l >= binStart && l < binEnd).length,
    };
  });
};

const LatencyChart = ({ data }) => {
  const histogramData = generateHistogramData(data.map((d) => d.latency));

  return (
    <Card sx={{ height: 400, p: 2 }}>
      <Typography variant="h6" gutterBottom>
        Latency Distribution
      </Typography>
      <ResponsiveContainer width="100%" height="90%">
        <BarChart data={histogramData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            dataKey="range"
            angle={-45}
            height={65}
            textAnchor="end"
            tick={{ fontSize: 12 }}
          />
          <YAxis />
          <Tooltip />
          <Bar dataKey="count" fill="#8884d8" />
        </BarChart>
      </ResponsiveContainer>
    </Card>
  );
};

const SummaryCards = ({ report, mode }) => {
  const { summary } = report;

  const formatThroughput = (mbps) => {
    if (mbps < 1) {
      return `${(mbps * 1024).toFixed(2)} KB/s`;
    }
    return `${mbps.toFixed(1)} MB/s`;
  };

  const formatKB = (kb) => {
    if (kb < 1024) return `${kb.toLocaleString()} KB`;
    return `${(kb / 1024).toFixed(1)} MB`;
  };

  return (
    <Grid container spacing={3} sx={{ mb: 4 }}>
      <Grid item xs={12} md={2}>
        <Card
          sx={{
            p: 2,
            height: 200,
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
            background: `linear(135deg, ${colorTheme[mode].primary}, ${colorTheme[mode].secondary})`,
            color: mode === "dark" ? "#ffffff" : "#000000",
          }}
        >
          <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
            <CircularProgress
              variant="determinate"
              value={
                (summary.successful_requests / summary.total_requests) * 100
              }
              size={60}
              thickness={4}
              sx={{ color: colorTheme[mode].warning }}
            />
            <Box>
              <Typography variant="h6">Success Rate</Typography>
              <Typography variant="h3">
                {(
                  (summary.successful_requests / summary.total_requests) *
                  100
                ).toFixed(1)}
                %
              </Typography>
            </Box>
          </Box>
        </Card>
      </Grid>

      <Grid item xs={12} md={2}>
        <Card
          sx={{
            p: 2,
            height: 200,
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
          }}
        >
          <Box>
            <Typography variant="h6">Total Requests</Typography>
            <Typography variant="h3">{summary.total_requests}</Typography>
            <Box sx={{ display: "flex", gap: 2, mt: 1 }}>
              <Typography color="success.main">
                +{summary.successful_requests}
              </Typography>
              <Typography color="error.main">
                -{summary.failed_requests}
              </Typography>
            </Box>
          </Box>
        </Card>
      </Grid>

      <Grid item xs={12} md={2}>
        <Card
          sx={{
            p: 2,
            height: 200,
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
          }}
        >
          <Box>
            <Typography variant="h6">Data Transferred</Typography>
            <Typography variant="h3">
              {formatKB(summary.total_data_kb)}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {summary.total_data_kb.toLocaleString()} KB
            </Typography>
            <Typography variant="h6">File Sizes</Typography>
            <Typography variant="body2" color="text.secondary">
              {report.config.FilesizePolicies.map((policy, index) => (
                <span key={index}>
                  {formatSize(policy.size)} ({policy.percent}%)
                  {index < report.config.FilesizePolicies.length - 1
                    ? ", "
                    : ""}
                </span>
              ))}
            </Typography>
          </Box>
        </Card>
      </Grid>

      <Grid item xs={12} md={2}>
        <Card
          sx={{
            p: 2,
            height: 200,
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
          }}
        >
          <Box>
            <Typography variant="h6">Avg Throughput</Typography>
            <Typography variant="h3">
              {formatThroughput(summary.avg_throughput_mbps)}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Peak: {formatThroughput(summary.peak_throughput_mbps)}
            </Typography>
          </Box>
        </Card>
      </Grid>

      <Grid item xs={12} md={2}>
        <Card
          sx={{
            p: 2,
            height: 200,
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
          }}
        >
          <Box>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 1 }}>
              {report.config.Type.toLowerCase() === "upload" ? (
                <UploadIcon
                  fontSize="large"
                  sx={{
                    color:
                      mode === "dark"
                        ? colorTheme.dark.warning
                        : colorTheme.light.primary,
                  }}
                />
              ) : (
                <DownloadIcon
                  fontSize="large"
                  sx={{
                    color:
                      mode === "dark"
                        ? colorTheme.dark.warning
                        : colorTheme.light.primary,
                  }}
                />
              )}
              <Typography variant="h6">Protocol</Typography>
            </Box>
            <Typography variant="h3" sx={{ textTransform: "uppercase" }}>
              {report.config.Protocol}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Direction: {report.config.Type}
            </Typography>
          </Box>
        </Card>
      </Grid>

      <Grid item xs={12} md={2}>
        <Card
          sx={{
            p: 2,
            height: 200,
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
          }}
        >
          <Box>
            <Typography variant="h6">Latency</Typography>
            <Typography variant="h3">
              {summary.avg_latency_ms.toFixed(1)}ms
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {summary.min_latency_ms.toFixed(0)}-
              {summary.max_latency_ms.toFixed(0)}ms
            </Typography>
          </Box>
        </Card>
      </Grid>
    </Grid>
  );
};

const PercentileChart = ({ percentiles }) => (
  <Card sx={{ p: 2, height: 300 }}>
    <Typography variant="h6" gutterBottom>
      Latency Percentiles
    </Typography>
    <ResponsiveContainer width="100%" height="80%">
      <BarChart
        data={[
          { name: "P25", value: percentiles.p25 },
          { name: "P50", value: percentiles.p50 },
          { name: "P75", value: percentiles.p75 },
          { name: "P90", value: percentiles.p90 },
          { name: "P95", value: percentiles.p95 },
          { name: "P99", value: percentiles.p99 },
        ]}
      >
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="name" />
        <YAxis unit="ms" />
        <Tooltip formatter={(value) => `${value.toFixed(1)}ms`} />
        <Bar dataKey="value" fill="#8884d8" />
      </BarChart>
    </ResponsiveContainer>
  </Card>
);

const ErrorDistribution = ({ errors }) => (
  <Card sx={{ p: 2, mt: 3 }}>
    <Typography variant="h6" gutterBottom>
      Error Distribution
    </Typography>
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Error Type</TableCell>
            <TableCell align="right">Count</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {Object.entries(errors).map(([error, count]) => (
            <TableRow key={error}>
              <TableCell>{error}</TableCell>
              <TableCell align="right">{count}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  </Card>
);

const ThroughputCard = ({ summary, fileSize }) => {
  const dataSizeKB = (summary.throughput_rps * fileSize).toFixed(1);

  return (
    <Card sx={{ p: 2, mt: 3 }}>
      <Typography variant="h6" gutterBottom>
        Data Throughput
      </Typography>
      <Box sx={{ display: "flex", alignItems: "baseline" }}>
        <Typography variant="h4" sx={{ mr: 1 }}>
          {dataSizeKB}KB/s
        </Typography>
        <Typography color="text.secondary">
          ({summary.throughput_rps.toFixed(1)} files/s)
        </Typography>
      </Box>
      <LinearProgress
        variant="determinate"
        value={Math.min(dataSizeKB, 100)}
        sx={{ height: 4, mt: 1 }}
      />
    </Card>
  );
};

const FilesOverTimeChart = ({ timeSeries }) => {
  const chartData = useMemo(() => {
    if (!timeSeries) return [];
    return timeSeries.map((entry) => ({
      time: entry.timestamp,
      success: entry.requests - entry.errors,
      errors: entry.errors,
      total: entry.requests,
    }));
  }, [timeSeries]);

  return (
    <Card sx={{ height: 300, p: 2, mt: 3 }}>
      <Typography variant="h6" gutterBottom>
        Request Success vs Errors Over Time
      </Typography>
      <ResponsiveContainer width="100%" height="80%">
        <AreaChart data={chartData} stackOffset="expand">
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            dataKey="time"
            tickFormatter={(time) => new Date(time).toLocaleTimeString()}
          />
          <YAxis />
          <Tooltip formatter={(value) => `${value} requests`} />
          <Legend />
          <Area
            type="monotone"
            dataKey="success"
            name="Successful Requests"
            stackId="1"
            fill="#82ca9d"
            stroke="#82ca9d"
          />
          <Area
            type="monotone"
            dataKey="errors"
            name="Failed Requests"
            stackId="1"
            fill="#ff7300"
            stroke="#ff7300"
          />
        </AreaChart>
      </ResponsiveContainer>
    </Card>
  );
};

const SlowRequestsTable = ({ latencies }) => {
  const sortedLatencies = useMemo(
    () => [...latencies].sort((a, b) => b - a).slice(0, 5),
    [latencies]
  );

  return (
    <Card sx={{ mt: 3, p: 2 }}>
      <Typography variant="h6" gutterBottom>
        Top 5 Slowest Requests
      </Typography>
      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Rank</TableCell>
              <TableCell align="right">Latency</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {sortedLatencies.map((latency, index) => (
              <TableRow key={index}>
                <TableCell>#{index + 1}</TableCell>
                <TableCell align="right">{formatDuration(latency)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Card>
  );
};

const TransfersTable = ({ timeSeries }) => {
  return (
    <Card sx={{ mt: 3, p: 2 }}>
      <Typography variant="h6" gutterBottom>
        Detailed Transfers
      </Typography>
      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Timestamp</TableCell>
              <TableCell align="right">Requests</TableCell>
              <TableCell align="right">Success</TableCell>
              <TableCell align="right">Errors</TableCell>
              <TableCell align="right">Data</TableCell>
              <TableCell align="right">Throughput</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {timeSeries.map((entry, index) => (
              <TableRow key={index}>
                <TableCell>
                  {new Date(entry.timestamp).toLocaleTimeString()}
                </TableCell>
                <TableCell align="right">{entry.requests}</TableCell>
                <TableCell align="right" sx={{ color: "success.main" }}>
                  {entry.requests - entry.errors}
                </TableCell>
                <TableCell align="right" sx={{ color: "error.main" }}>
                  {entry.errors}
                </TableCell>
                <TableCell align="right">
                  {entry.data_transferred_kb < 1024
                    ? `${entry.data_transferred_kb.toFixed(2)} KB`
                    : `${(entry.data_transferred_kb / 1024).toFixed(2)} MB`}
                </TableCell>
                <TableCell align="right">
                  {entry.throughput_mbps < 1
                    ? `${(entry.throughput_mbps * 1024).toFixed(2)} KB/s`
                    : `${entry.throughput_mbps.toFixed(2)} MB/s`}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Card>
  );
};

const colorTheme = {
  dark: {
    primary: "#003f5c",
    secondary: "#58508d",
    accent1: "#bc5090",
    accent2: "#ff6361",
    warning: "#ffa600",
    background: "#121212",
    text: "#ffffff",
  },
  light: {
    primary: "#003f5c",
    secondary: "#58508d",
    accent1: "#bc5090",
    accent2: "#ff6361",
    warning: "#ffa600",
    background: "#f5f5f5",
    text: "#121212",
  },
};

const getTheme = (mode) =>
  createTheme({
    palette: {
      mode,
      primary: { main: colorTheme[mode].primary },
      secondary: { main: colorTheme[mode].secondary },
      error: { main: colorTheme[mode].accent2 },
      warning: { main: colorTheme[mode].warning },
      background: { default: colorTheme[mode].background },
    },
    components: {
      MuiCard: {
        styleOverrides: {
          root: {
            backgroundColor: mode === "dark" ? "#1e1e1e" : "#ffffff",
            transition: "all 0.3s ease",
            "&:hover": { transform: "translateY(-2px)" },
          },
        },
      },
    },
  });

const ReportViewer = () => {
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const [errorMessage, setErrorMessage] = useState(null);
  const [mode, setMode] = useState("dark");

  const validateReportStructure = (data) => {
    const requiredFields = [
      "config",
      "summary",
      "latencies",
      "throughputs",
      "errors",
      "timestamp",
    ];

    const missingFields = requiredFields.filter((field) => !(field in data));
    if (missingFields.length > 0) {
      throw new Error(
        `Invalid report format. Missing fields: ${missingFields.join(", ")}`
      );
    }

    // Validate nested structure
    const configFields = ["Protocol", "Host", "Port", "FilesizePolicies"];
    const missingConfig = configFields.filter((f) => !(f in data.config));
    if (missingConfig.length > 0) {
      throw new Error(`Missing config fields: ${missingConfig.join(", ")}`);
    }

    // Validate latencies array
    if (!Array.isArray(data.latencies) || data.latencies.some(isNaN)) {
      throw new Error("Latencies must be an array of numbers");
    }

    // Validate summary structure
    const summaryFields = [
      "total_requests",
      "successful_requests",
      "failed_requests",
      "avg_latency_ms",
      "min_latency_ms",
      "max_latency_ms",
    ];
    const missingSummary = summaryFields.filter((f) => !(f in data.summary));
    if (missingSummary.length > 0) {
      throw new Error(`Missing summary fields: ${missingSummary.join(", ")}`);
    }

    // Validate timestamp format
    if (isNaN(new Date(data.timestamp).getTime())) {
      throw new Error("Invalid timestamp format");
    }

    return true;
  };

  const toggleTheme = () =>
    setMode((prev) => (prev === "light" ? "dark" : "light"));

  const { getRootProps, getInputProps } = useDropzone({
    accept: {
      "application/json": [".json"],
    },
    onDrop: (acceptedFiles) => {
      handleFileUpload({ target: { files: acceptedFiles } });
      setDragActive(false);
    },
    multiple: false,
    onDragEnter: () => setDragActive(true),
    onDragLeave: () => setDragActive(false),
  });

  const handleFileUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    setLoading(true);
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = JSON.parse(e.target.result);
        setReport(data);
        validateReportStructure(data);
      } catch (err) {
        setReport(null);
        const message =
          err instanceof SyntaxError
            ? "Invalid JSON file. Please check the file format."
            : `Invalid report: ${err.message}`;
        setErrorMessage(message);
        console.error("Invalid report file:", err);
      } finally {
        setTimeout(() => setErrorMessage(null), 5000);
        setLoading(false);
      }
    };
    reader.readAsText(file);
  };

  const latencyData = useMemo(
    () =>
      report?.latencies.map((latency, index) => ({
        request: index + 1,
        latency,
        timestamp: report.timestamp,
      })) || [],
    [report]
  );

  if (!report) {
    return (
      <ThemeProvider theme={getTheme(mode)}>
        <CssBaseline />
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            minHeight: "80vh",
            p: 3,
          }}
        >
          <Box
            {...getRootProps()}
            sx={{
              width: "100%",
              maxWidth: 600,
              border: `3px dashed ${colorTheme[mode].primary}`,
              borderRadius: "15px",
              p: 6,
              textAlign: "center",
              background:
                mode === "dark" ? "rgba(0,63,92,0.1)" : "rgba(0,63,92,0.05)",
              cursor: "pointer",
              "&:hover": {
                borderColor: colorTheme[mode].secondary,
                background:
                  mode === "dark"
                    ? "rgba(88,80,141,0.1)"
                    : "rgba(88,80,141,0.05)",
              },
            }}
          >
            <input {...getInputProps()} />
            <CloudUploadIcon
              sx={{
                fontSize: 64,
                color: colorTheme[mode].primary,
                animation: dragActive ? "bounce 1s infinite" : "none",
              }}
            />
            <Typography
              variant="h5"
              gutterBottom
              sx={{ color: colorTheme[mode].text }}
            >
              Drag and Drop Test Report
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mb: 2 }}>
              or click to browse files
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Supported format: JSON (.json)
            </Typography>
          </Box>

          {loading && (
            <Box sx={{ width: "100%", maxWidth: 600, mt: 4 }}>
              <LinearProgress />
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Analyzing report...
              </Typography>
            </Box>
          )}

          {errorMessage && (
            <Paper
              elevation={3}
              sx={{ mt: 2, p: 2, bgcolor: "error.main", color: "white" }}
            >
              <Typography variant="body1" align="center">
                {errorMessage}
              </Typography>
            </Paper>
          )}
        </Box>
      </ThemeProvider>
    );
  }

  return (
    <ThemeProvider theme={getTheme(mode)}>
      <CssBaseline />
      <Box sx={{ p: 4 }}>
        <Box
          sx={{
            mb: 4,
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            p: 3,
            borderRadius: "15px",
            background: `linear(135deg, ${colorTheme[mode].primary} 0%, ${colorTheme[mode].secondary} 100%)`,
            boxShadow: 3,
            transition: "all 0.3s ease",
          }}
        >
          <Box sx={{ display: "flex", alignItems: "center", gap: 3 }}>
            <img
              src="./aionyx-logo.png"
              alt="Aionyx - MFT Runner Viewer"
              style={{
                height: "100px",
                transition: "filter 0.3s ease",
              }}
            />
            <Typography
              variant="h4"
              sx={{
                fontWeight: "bold",
                letterSpacing: "1.5px",
                textShadow: "2px 2px 4px rgba(0,0,0,0.2)",
                fontFamily: "ui-sans-serif, sans-serif",
                display: "flex",
                alignItems: "center",
                color: colorTheme[mode].text,
              }}
            >
              <span style={{ color: mode === "dark" ? "#fff" : "#000" }}>
                Ai
              </span>
              <span
                style={{
                  color: mode === "dark" ? "#8884d8" : "rgb(79 70 229)",
                }}
              >
                onyx
              </span>
              <span style={{ marginLeft: "8px" }}> - MFT Runner Viewer</span>
            </Typography>
          </Box>
          <Box sx={{ display: "flex", alignItems: "center", gap: 3 }}>
            <IconButton
              onClick={toggleTheme}
              sx={{
                color: mode === "dark" ? "#ffffff" : colorTheme[mode].primary,
                "&:hover": {
                  backgroundColor: alpha("#ffffff", 0.1),
                },
              }}
            >
              {mode === "dark" ? <Brightness7Icon /> : <Brightness4Icon />}
            </IconButton>
            <Button
              variant="contained"
              color="warning"
              onClick={() => setReport(null)}
              startIcon={<RefreshIcon />}
              sx={{
                fontSize: "1.1rem",
                py: 1.5,
                px: 4,
                borderRadius: "8px",
                boxShadow: 3,
                background: `linear(45deg, ${colorTheme[mode].accent1}, ${colorTheme[mode].accent2})`,
                "&:hover": {
                  transform: "translateY(-2px)",
                  boxShadow: 6,
                },
                transition: "all 0.3s ease",
              }}
            >
              New Report
            </Button>
          </Box>
        </Box>

        <SummaryCards report={report} mode={mode} />

        <Grid container spacing={3}>
          <Grid item xs={12} md={8}>
            <Card sx={{ p: 2, mb: 3, height: 400 }}>
              <Typography variant="h6" gutterBottom>
                Throughput Over Time
              </Typography>
              <ResponsiveContainer width="100%" height="90%">
                <LineChart data={report.time_series}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis
                    dataKey="timestamp"
                    tickFormatter={(time) =>
                      new Date(time).toLocaleTimeString([], {
                        hour: "2-digit",
                        minute: "2-digit",
                        second: "2-digit",
                      })
                    }
                  />
                  <YAxis
                    yAxisId="left"
                    label={{
                      value: "Requests/s",
                      angle: -90,
                      position: "insideLeft",
                    }}
                  />
                  <YAxis
                    yAxisId="right"
                    orientation="right"
                    label={{
                      value: "Data Rate",
                      angle: 90,
                      position: "insideRight",
                    }}
                  />
                  <Tooltip
                    formatter={(value, name) => {
                      if (name === "Throughput (RPS)") {
                        return [value.toFixed(1), name];
                      }
                      if (name === "Data Throughput") {
                        return value < 1
                          ? [`${(value * 1024).toFixed(2)} KB/s`, name]
                          : [`${value.toFixed(2)} MB/s`, name];
                      }
                      return [value, name];
                    }}
                  />
                  <Legend
                    verticalAlign="top"
                    wrapperStyle={{ paddingBottom: 20 }}
                  />
                  <Line
                    yAxisId="left"
                    type="monotone"
                    dataKey="throughput_rps"
                    stroke="#8884d8"
                    name="Requests/sec"
                    strokeWidth={2}
                    dot={false}
                  />
                  <Line
                    yAxisId="right"
                    type="monotone"
                    dataKey="throughput_mbps"
                    stroke="#82ca9d"
                    name="Data Throughput"
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            </Card>

            <LatencyChart data={latencyData} />
            <FilesOverTimeChart timeSeries={report.time_series} />
            <TransfersTable timeSeries={report.time_series} />
          </Grid>

          <Grid item xs={12} md={4}>
            <PercentileChart percentiles={report.summary.percentiles} />
            <ErrorDistribution errors={report.summary.error_distribution} />
            <SlowRequestsTable latencies={report.latencies} />
            <Card sx={{ mt: 3, p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Test Configuration
              </Typography>
              <TableContainer>
                <Table size="small">
                  <TableBody>
                    <TableRow>
                      <TableCell>Protocol</TableCell>
                      <TableCell>{report.config.Protocol}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell>Type</TableCell>
                      <TableCell>{report.config.Type}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell>Workers</TableCell>
                      <TableCell>{report.config.NumClients}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell>Total Requests</TableCell>
                      <TableCell>{report.summary.total_requests}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell>File Size</TableCell>
                      <TableCell>
                        {report.config.FilesizePolicies.map((policy, index) => (
                          <span key={index}>
                            {formatSize(policy.size)} ({policy.percent}%)
                            {index < report.config.FilesizePolicies.length - 1
                              ? ", "
                              : ""}
                          </span>
                        ))}
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </TableContainer>
            </Card>
          </Grid>
        </Grid>
      </Box>
    </ThemeProvider>
  );
};

export default ReportViewer;
