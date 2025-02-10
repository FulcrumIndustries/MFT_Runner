import React, { useState, useEffect } from "react";
import "./index.css";
import {
  Container,
  Grid,
  Select,
  MenuItem,
  TextField,
  Button,
  LinearProgress,
  Typography,
  Box,
  Card,
  CardContent,
  CardHeader,
  Divider,
  IconButton,
  Alert,
  Snackbar,
  useTheme,
  alpha,
  Stack,
  Chip,
  Tooltip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Dialog,
  DialogTitle,
  DialogContent,
  CircularProgress,
  Paper,
} from "@mui/material";
import {
  PlayArrow as StartIcon,
  Stop as StopIcon,
  Add as AddIcon,
  Refresh as RefreshIcon,
  Speed as SpeedIcon,
  Memory as MemoryIcon,
  Delete as DeleteIcon,
  UploadFile as UploadIcon,
  DownloadForOffline as DownloadIcon,
  Article as ArticleIcon,
} from "@mui/icons-material";
import CampaignEditor from "./CampaignEditor";
import { getTestHistory } from './api';

function App() {
  const theme = useTheme();
  const [campaigns, setCampaigns] = useState([]);
  const [selectedCampaign, setSelectedCampaign] = useState("");
  const [editCampaign, setEditCampaign] = useState(null);
  const [showEditor, setShowEditor] = useState(false);
  const [numClients, setNumClients] = useState(1);
  const [numRequests, setNumRequests] = useState(10);
  const [running, setRunning] = useState(false);
  const [progress, setProgress] = useState(0);
  const [results, setResults] = useState([]);
  const [errors, setErrors] = useState([]);
  const [report, setReport] = useState(null);
  const [intervalId, setIntervalId] = useState(null);
  const [notification, setNotification] = useState({
    open: false,
    message: "",
    severity: "info",
  });
  const [campaignDetails, setCampaignDetails] = useState(null);
  const [successCount, setSuccessCount] = useState(0);
  const [failureCount, setFailureCount] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const [testHistory, setTestHistory] = useState([]);
  const [currentTest, setCurrentTest] = useState(null);
  const [selectedTest, setSelectedTest] = useState(null);
  const [, setTick] = useState(0); // Used to force re-render for elapsed time
  const [selectedLogs, setSelectedLogs] = useState(null);
  const [showLogs, setShowLogs] = useState(false);
  const [liveLogs, setLiveLogs] = useState([]);
  const [status, setStatus] = useState(null);
  const [cliConnected, setCliConnected] = useState(false);
  const [activeWorkers, setActiveWorkers] = useState(0);
  const [throughput, setThroughput] = useState(0);
  const [testStatus, setTestStatus] = useState(null);
  const [realtimeLogs, setRealtimeLogs] = useState([]);
  const [eventSource, setEventSource] = useState(null);

  useEffect(() => {
    // Load campaigns on mount
    fetch("/api/campaigns")
      .then((res) => {
        if (!res.ok) throw new Error(`Backend unavailable (${res.status})`);
        return res.json();
      })
      .then((data) => {
        console.log("Loaded campaigns:", data);
        if (data && data.length > 0) {
          console.log(
            "First campaign structure:",
            JSON.stringify(data[0], null, 2)
          );
        }
        setCampaigns(data || []);
        setSelectedCampaign("");
      })
      .catch((err) => {
        console.error("Backend connection failed:", err);
        setErrors((prev) => [...prev, "Backend service unavailable"]);
        setNotification({
          open: true,
          message: "Cannot connect to backend service",
          severity: "error",
        });
        setCampaigns([]);
      });
  }, []);

  useEffect(() => {
    const es = new EventSource("http://localhost:8080/api/events", {
      withCredentials: true,
    });

    es.onopen = () => console.log("SSE connection established");
    es.onerror = (err) => console.error("SSE error:", err);

    es.onmessage = (event) => {
      try {
        console.log("SSE message received:", event.data);
        const data = JSON.parse(event.data);
        if (data.type === "status") {
          setTestHistory((prev) => {
            const existingIndex = prev.findIndex(
              (t) => t.testId === data.payload.testId
            );
            if (existingIndex > -1) {
              const updated = [...prev];
              updated[existingIndex] = data.payload;
              return updated;
            }
            return [...prev, data.payload];
          });
        }
      } catch (error) {
        console.error("Error parsing SSE data:", error);
      }
    };

    return () => es.close();
  }, []);

  // Add effect to clear current test when completed
  useEffect(() => {
    if (currentTest?.status === "completed") {
      setCurrentTest(null);
    }
  }, [currentTest]);

  const readLogs = async () => {
    try {
      const response = await fetch("/logs/latest.log");
      const text = await response.text();
      const entries = text
        .split("\n")
        .filter((line) => line.trim())
        .map((line) => {
          try {
            return JSON.parse(line);
          } catch {
            return null;
          }
        })
        .filter(Boolean);

      // Process status updates
      const statusEntries = entries.filter((e) => e.progress !== undefined);
      if (statusEntries.length > 0) {
        const latestStatus = statusEntries[statusEntries.length - 1];
        setProgress(latestStatus.progress);
        setSuccessCount(latestStatus.success);
        setFailureCount(latestStatus.failures);
      }

      // Process results
      const testResults = entries.filter((e) => e.type === "result");
      setResults(testResults);

      // Update test history from log entries
      const historyEntries = entries
        .filter((e) => e.type === "testStart" || e.type === "result")
        .reduce((acc, entry) => {
          if (entry.type === "testStart") {
            acc[entry.testId] = {
              id: entry.testId,
              campaign: entry.campaign,
              date: new Date(entry.timestamp * 1000),
              numClients: entry.numClients,
              numRequests: entry.numRequests,
              status: "Running",
            };
          } else if (entry.type === "result") {
            if (acc[entry.testId]) {
              acc[entry.testId] = {
                ...acc[entry.testId],
                success: entry.success,
                failures: entry.failures,
                duration: entry.duration,
                status: "Completed",
              };
            }
          }
          return acc;
        }, {});

      setTestHistory(Object.values(historyEntries).reverse());

      // Update live logs with raw entries
      const rawLogs = entries
        .filter((e) => e.type === "log")
        .map((e) => e.message);
      setLiveLogs(rawLogs.slice(-100)); // Keep last 100 entries
    } catch (err) {
      console.error("Error reading logs:", err);
    }
  };

  useEffect(() => {
    loadTestHistory();
    const interval = setInterval(loadTestHistory, 5000);
    return () => clearInterval(interval);
  }, []);

  const loadTestHistory = async () => {
    try {
      const response = await fetch("/api/tests/history");
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      setTestHistory(data || []);
    } catch (error) {
      console.error("Error loading test history:", error);
    }
  };

  const startTest = async () => {
    setRunning(true);
    setResults([]);
    setErrors([]);
    setReport(null);
    setSuccessCount(0);
    setFailureCount(0);
    setTotalCount(numClients * numRequests);

    try {
      const res = await fetch("/api/tests", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          campaign: selectedCampaign,
          workers: parseInt(numClients, 10),
          requests: parseInt(numRequests, 10),
          cli: false, // Flag for UI-initiated tests
        }),
      });

      if (!res.ok) {
        const error = await res.text();
        setErrors((prev) => [...prev, error]);
        setRunning(false);
        return;
      }

      // Add new test to history after successful start
      const responseData = await res.json();
      const newTest = {
        id: responseData.testId,
        campaign: selectedCampaign,
        date: new Date(),
        numClients: numClients,
        numRequests: numRequests,
        status: "Running",
      };
      setTestHistory((prev) => [newTest, ...prev]);
    } catch (err) {
      setErrors((prev) => [...prev, `Connection error: ${err.message}`]);
      setRunning(false);
    }
  };

  const stopTest = () => {
    fetch("/api/test/stop", { method: "POST" });
    setRunning(false);
    if (intervalId) clearInterval(intervalId);
  };

  const handleCreateCampaign = () => {
    setEditCampaign(null);
    setShowEditor(true);
  };

  const handleEditCampaign = (campaign) => {
    setEditCampaign(campaign);
    setShowEditor(true);
  };

  const refreshCampaigns = () => {
    fetch("/api/campaigns")
      .then((res) => res.json())
      .then((data) => setCampaigns(data));
  };

  const saveCampaign = (campaignData) => {
    const method = editCampaign ? "PUT" : "POST";
    const url = editCampaign
      ? `/api/campaigns/${editCampaign}`
      : "/api/campaigns";

    fetch(url, {
      method: method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(campaignData),
    }).then(() => {
      setShowEditor(false);
      refreshCampaigns();
    });
  };

  const showNotification = (message, severity = "info") => {
    setNotification({ open: true, message, severity });
  };

  const loadCampaignDetails = async (campaign) => {
    console.log("Original campaign path:", campaign);
    try {
      // Just use the campaign path as-is since it already includes 'Campaigns/'
      const res = await fetch(`/${campaign}`);
      const data = await res.json();
      console.log("Loaded campaign details:", data);
      setCampaignDetails(data);
    } catch (err) {
      console.error("Failed to load campaign details:", err);
    }
  };

  const handleCampaignSelect = (event) => {
    const campaign = event.target.value;
    setSelectedCampaign(campaign);
    if (campaign) {
      loadCampaignDetails(campaign);
    } else {
      setCampaignDetails(null);
    }
  };

  const handleDeleteTest = async (testId) => {
    try {
      await fetch(`/api/tests/${testId}`, {
        method: "DELETE",
      });
      loadTestHistory(); // Refresh the list after deletion
    } catch (error) {
      console.error("Error deleting test:", error);
    }
  };

  // Modify table rendering to show both current and historical tests
  const allTests = [
    ...(currentTest
      ? [
          {
            ...currentTest,
            id: currentTest.testId || currentTest.id,
            status: currentTest.status || "running",
          },
        ]
      : []),
    ...testHistory,
  ].filter((t) => t.status !== "idle");

  return (
    <>
      <Container maxWidth="xl" sx={{ py: 4 }}>
        {errors.length > 0 && (
          <Alert severity="error" sx={{ mb: 4 }}>
            Connection Error: {errors[errors.length - 1]}
          </Alert>
        )}

        <Box
          sx={{
            display: "flex",
            justifyContent: "space-between",
            mb: 4,
            alignItems: "center",
            backgroundColor: alpha(theme.palette.primary.main, 0.03),
            p: 3,
            borderRadius: 2,
            boxShadow: `inset 0 0 0 1px ${alpha(
              theme.palette.primary.main,
              0.1
            )}`,
          }}
        >
          <Stack
            direction="row"
            alignItems="center"
            spacing={2}
            sx={{ flex: 1 }}
          >
            <SpeedIcon
              sx={{ fontSize: 40, color: theme.palette.primary.main }}
            />
            <Typography
              variant="h4"
              sx={{
                fontWeight: 600,
                background: `linear-gradient(45deg, ${theme.palette.primary.main}, ${theme.palette.primary.dark})`,
                WebkitBackgroundClip: "text",
                WebkitTextFillColor: "transparent",
              }}
            >
              MFT Runner
            </Typography>
          </Stack>
        </Box>

        <Grid container spacing={4}>
          {campaigns.length === 0 && !errors.length && (
            <Grid item xs={12}>
              <Alert severity="warning">
                No campaigns found. Create one to get started.
              </Alert>
            </Grid>
          )}
          <Grid item xs={12} md={6}>
            <Card
              elevation={0}
              sx={{
                borderRadius: 3,
                border: `1px solid ${alpha(theme.palette.primary.main, 0.1)}`,
                background: theme.palette.background.paper,
                "&:hover": {
                  boxShadow: theme.shadows[4],
                },
                transition: "box-shadow 0.3s ease-in-out",
              }}
            >
              <CardHeader
                title={
                  <Stack direction="row" spacing={2} alignItems="center">
                    <MemoryIcon color="primary" />
                    <Typography variant="h6">Test Configuration</Typography>
                  </Stack>
                }
                action={
                  <IconButton
                    onClick={refreshCampaigns}
                    sx={{
                      "&:hover": {
                        backgroundColor: alpha(theme.palette.primary.main, 0.1),
                      },
                    }}
                  >
                    <RefreshIcon />
                  </IconButton>
                }
              />
              <Divider />
              <CardContent>
                <Typography
                  variant="subtitle1"
                  sx={{ mb: 1, color: theme.palette.text.secondary }}
                >
                  Campaign Configuration
                </Typography>
                <Stack direction="row" spacing={2} sx={{ mb: 3 }}>
                  <Select
                    fullWidth
                    value={selectedCampaign || ""}
                    onChange={handleCampaignSelect}
                    displayEmpty
                    sx={{
                      "& .MuiOutlinedInput-notchedOutline": {
                        borderColor: alpha(theme.palette.primary.main, 0.2),
                      },
                      "&:hover .MuiOutlinedInput-notchedOutline": {
                        borderColor: theme.palette.primary.main,
                      },
                    }}
                  >
                    <MenuItem value="">Select Campaign</MenuItem>
                    {campaigns?.map((c, index) => (
                      <MenuItem key={`${c}-${index}`} value={c}>
                        {c.replace(".json", "").replace("campaigns/", "")}
                      </MenuItem>
                    ))}
                  </Select>
                  <Button
                    variant="outlined"
                    color="primary"
                    startIcon={<AddIcon />}
                    onClick={handleCreateCampaign}
                    sx={{
                      whiteSpace: "nowrap",
                      px: 3,
                      borderWidth: 2,
                      "&:hover": {
                        borderWidth: 2,
                      },
                    }}
                  >
                    New Campaign
                  </Button>
                </Stack>

                <TextField
                  fullWidth
                  label="Number of Clients"
                  type="number"
                  value={numClients}
                  onChange={(e) => setNumClients(e.target.value)}
                  sx={{ mb: 3 }}
                />

                <TextField
                  fullWidth
                  label="Number of Requests"
                  type="number"
                  value={numRequests}
                  onChange={(e) => setNumRequests(e.target.value)}
                  sx={{ mb: 3 }}
                />

                <Box sx={{ display: "flex", gap: 2 }}>
                  <Button
                    variant="contained"
                    color="primary"
                    onClick={startTest}
                    disabled={running || !selectedCampaign}
                    startIcon={<StartIcon />}
                    sx={{
                      flex: 1,
                      borderRadius: 2,
                      py: 1.5,
                      background: `linear-gradient(45deg, ${theme.palette.success.main}, ${theme.palette.success.dark})`,
                      "&:hover": {
                        background: `linear-gradient(45deg, ${theme.palette.success.dark}, ${theme.palette.success.main})`,
                      },
                      "&:disabled": {
                        background: theme.palette.grey[300],
                      },
                    }}
                  >
                    Start Test
                  </Button>

                  <Button
                    variant="contained"
                    color="error"
                    onClick={stopTest}
                    disabled={!running}
                    startIcon={<StopIcon />}
                    sx={{
                      flex: 1,
                      borderRadius: 2,
                      py: 1.5,
                      background: `linear-gradient(45deg, ${theme.palette.error.main}, ${theme.palette.error.dark})`,
                      "&:hover": {
                        background: `linear-gradient(45deg, ${theme.palette.error.dark}, ${theme.palette.error.main})`,
                      },
                      "&:disabled": {
                        background: theme.palette.grey[300],
                      },
                    }}
                  >
                    Stop
                  </Button>
                </Box>
              </CardContent>
            </Card>
          </Grid>

          {/* Campaign Details Card */}
          {campaignDetails && (
            <Card sx={{ mt: 4 }}>
              <CardHeader
                title={
                  <Stack direction="row" spacing={2} alignItems="center">
                    <MemoryIcon color="primary" />
                    <Typography variant="h6">Campaign Details</Typography>
                  </Stack>
                }
              />
              <CardContent>
                <Grid container spacing={3}>
                  <Grid item xs={12} md={6}>
                    <TableContainer>
                      <Table size="small">
                        <TableBody>
                          <TableRow>
                            <TableCell component="th">Protocol</TableCell>
                            <TableCell>
                              <Chip
                                label={campaignDetails.Protocol}
                                color="primary"
                                size="small"
                              />
                            </TableCell>
                          </TableRow>
                          <TableRow>
                            <TableCell component="th">Transfer Type</TableCell>
                            <TableCell>
                              <Chip
                                label={campaignDetails.Type}
                                color="secondary"
                                size="small"
                              />
                            </TableCell>
                          </TableRow>
                          <TableRow>
                            <TableCell component="th">Timeout</TableCell>
                            <TableCell>
                              {campaignDetails.Timeout} seconds
                              <LinearProgress
                                variant="determinate"
                                value={Math.min(campaignDetails.Timeout, 30)}
                                sx={{ mt: 1, height: 4 }}
                              />
                            </TableCell>
                          </TableRow>
                        </TableBody>
                      </Table>
                    </TableContainer>
                  </Grid>

                  <Grid item xs={12} md={6}>
                    <Card variant="outlined" sx={{ p: 2 }}>
                      <Typography variant="subtitle2" gutterBottom>
                        File Size Distribution
                      </Typography>
                      {campaignDetails.FilesizePolicies.map((policy, i) => (
                        <Box key={i} sx={{ mb: 2 }}>
                          <Stack
                            direction="row"
                            spacing={1}
                            alignItems="center"
                          >
                            <Box
                              sx={{
                                width: 8,
                                height: 8,
                                borderRadius: "50%",
                                backgroundColor: theme.palette.primary.main,
                              }}
                            />
                            <Typography variant="body2">
                              {policy.Size} {policy.Unit} ({policy.Percent}%)
                            </Typography>
                          </Stack>
                          <LinearProgress
                            variant="determinate"
                            value={policy.Percent}
                            sx={{ height: 4, mt: 0.5 }}
                          />
                        </Box>
                      ))}
                    </Card>
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          )}

          <Grid item xs={12}>
            <Card sx={{ mt: 4, bgcolor: "background.paper" }}>
              <CardHeader
                title={
                  <Stack direction="row" spacing={2} alignItems="center">
                    <ArticleIcon color="primary" />
                    <Typography variant="h6">
                      Real-time Execution Logs
                    </Typography>
                  </Stack>
                }
                sx={{
                  borderBottom: `1px solid ${alpha(
                    theme.palette.divider,
                    0.1
                  )}`,
                }}
              />
              <CardContent sx={{ p: 0 }}>
                <Box
                  sx={{
                    backgroundColor: theme.palette.grey[900],
                    borderRadius: 2,
                    p: 2,
                    height: 100,
                    width: "100%",
                    overflow: "auto",
                    fontFamily: "monospace",
                    color: theme.palette.common.white,
                    "& pre": {
                      margin: 0,
                      whiteSpace: "pre-wrap",
                      wordBreak: "break-word",
                    },
                  }}
                >
                  <pre>
                    {realtimeLogs.map((log, index) => (
                      <Box
                        key={index}
                        component="div"
                        sx={{
                          color: log.includes("ERROR")
                            ? "error.main"
                            : log.includes("WARN")
                            ? "warning.main"
                            : "success.main",
                          py: 0.5,
                          animation: "fadeIn 0.3s ease",
                          fontSize: "0.875rem",
                        }}
                      >
                        {log}
                      </Box>
                    ))}
                  </pre>
                </Box>
              </CardContent>
            </Card>
            <Card
              elevation={0}
              sx={{
                borderRadius: 3,
                border: `1px solid ${alpha(theme.palette.primary.main, 0.1)}`,
                background: theme.palette.background.paper,
                mt: 4,
              }}
            >
              <CardHeader
                title="Tests"
                titleTypographyProps={{ variant: "h5" }}
                action={
                  running && (
                    <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
                      <LinearProgress
                        sx={{ width: 100 }}
                        variant="indeterminate"
                        color="primary"
                      />
                      <Typography variant="body2" color="primary">
                        Test in progress...
                      </Typography>
                    </Box>
                  )
                }
              />
              <Divider />
              <CardContent>
                <Typography variant="h5" sx={{ mb: 2 }}>
                  Active Tests
                </Typography>
                <TestHistoryTable filter="active" testHistory={testHistory} />

                <Typography variant="h5" sx={{ mt: 4, mb: 2 }}>
                  Test History
                </Typography>
                <TestHistoryTable
                  filter="completed"
                  testHistory={testHistory}
                />
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} md={4}>
            <Card
              elevation={0}
              sx={{
                borderRadius: 3,
                border: `1px solid ${alpha(theme.palette.primary.main, 0.1)}`,
                background: theme.palette.background.paper,
              }}
            ></Card>
          </Grid>
        </Grid>

        <CampaignEditor
          open={showEditor}
          campaign={editCampaign}
          onClose={() => setShowEditor(false)}
          onSave={saveCampaign}
        />

        <Snackbar
          open={notification.open}
          autoHideDuration={6000}
          onClose={() => setNotification((prev) => ({ ...prev, open: false }))}
        >
          <Alert
            severity={notification.severity}
            variant="filled"
            sx={{ width: "100%" }}
            onClose={() =>
              setNotification((prev) => ({ ...prev, open: false }))
            }
          >
            {notification.message}
          </Alert>
        </Snackbar>

        <Dialog
          open={showLogs}
          onClose={() => setShowLogs(false)}
          maxWidth="md"
          fullWidth
        >
          <DialogTitle>Test Execution Logs</DialogTitle>
          <DialogContent>
            <Box
              sx={{
                p: 2,
                backgroundColor: theme.palette.background.default,
                borderRadius: 2,
                fontFamily: "monospace",
                whiteSpace: "pre-wrap",
                maxHeight: "60vh",
                overflow: "auto",
              }}
            >
              {realtimeLogs.join("\n")}
            </Box>
          </DialogContent>
        </Dialog>

        <style>
          {`
          @keyframes fadeIn {
            from { opacity: 0; transform: translateY(5px); }
            to { opacity: 1; transform: translateY(0); }
          }
          @keyframes pulse {
            0% { transform: scale(1); opacity: 1; }
            50% { transform: scale(1.2); opacity: 0.5; }
            100% { transform: scale(1); opacity: 1; }
          }
          `}
        </style>
      </Container>

      <Container maxWidth="xl">{/* Main content */}</Container>
    </>
  );
}

const StatItem = ({ label, value, color }) => (
  <Box
    sx={{
      p: 2,
      borderRadius: 2,
      backgroundColor: (theme) => alpha(theme.palette.primary.main, 0.05),
      border: (theme) => `1px solid ${alpha(theme.palette.primary.main, 0.1)}`,
    }}
  >
    <Typography
      variant="body2"
      sx={{ color: (theme) => theme.palette.text.secondary }}
    >
      {label}
    </Typography>
    <Typography
      variant="h6"
      sx={{
        color: (theme) => color || theme.palette.primary.main,
        fontWeight: 500,
      }}
    >
      {value}
    </Typography>
  </Box>
);

const TestHistoryTable = ({ filter, testHistory }) => {
  const filteredTests = (testHistory || []).filter((test) =>
    filter === "active"
      ? test?.status !== "completed"
      : test?.status === "completed"
  );

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Test ID</TableCell>
            <TableCell>Date</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Campaign</TableCell>
            <TableCell>Protocol</TableCell>
            <TableCell>Direction</TableCell>
            <TableCell>Workers</TableCell>
            <TableCell>Requests</TableCell>
            <TableCell>Success</TableCell>
            <TableCell>Failures</TableCell>
            <TableCell>Duration</TableCell>
            <TableCell>Throughput</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {filteredTests.map((test) => (
            <TableRow key={test?.testId}>
              <TableCell>{test?.testId || "N/A"}</TableCell>
              <TableCell>
                {test?.date ? new Date(test.date).toLocaleString() : "N/A"}
              </TableCell>
              <TableCell>
                <Chip
                  label={test?.status || "unknown"}
                  color={test?.status === "completed" ? "success" : "primary"}
                />
              </TableCell>
              <TableCell>{test?.campaign || "N/A"}</TableCell>
              <TableCell>{test?.protocol || "N/A"}</TableCell>
              <TableCell>{test?.direction || "N/A"}</TableCell>
              <TableCell>{test?.numClients || 0}</TableCell>
              <TableCell>{test?.numRequests || 0}</TableCell>
              <TableCell>{test?.success || 0}</TableCell>
              <TableCell>{test?.failures || 0}</TableCell>
              <TableCell>{(test?.duration || 0).toFixed(2)}s</TableCell>
              <TableCell>{(test?.throughput || 0).toFixed(2)}/s</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default App;
