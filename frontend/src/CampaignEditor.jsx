import React, { useState, useEffect } from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, TextField, Button, Select, MenuItem, FormControl, InputLabel, IconButton, Box, Typography, Divider } from '@mui/material';
import { Add as AddIcon, Remove as RemoveIcon } from '@mui/icons-material';

export default function CampaignEditor({ open, campaign, onClose, onSave }) {
    const [formState, setFormState] = useState({
        name: '',
        Protocol: 'FTP',
        Type: 'Download',
        Hostname: '',
        Port: 21,
        Path: '/',
        Timeout: 5,
        FilesizePolicies: [{ Size: 1, Unit: 'KB', Percent: 100 }]
    });

    useEffect(() => {
        if (campaign) {
            // Convert numeric fields when loading existing campaign
            const convertedCampaign = {
                ...campaign,
                FilesizePolicies: campaign.FilesizePolicies.map(policy => ({
                    ...policy,
                    Size: Number(policy.Size),
                    Percent: Number(policy.Percent)
                }))
            };
            setFormState(convertedCampaign);
        }
    }, [campaign]);

    const handleChange = (e) => {
        setFormState({ ...formState, [e.target.name]: e.target.value });
    };

    const handleFilePolicyChange = (index, field, value) => {
        const newPolicies = [...formState.FilesizePolicies];
        newPolicies[index][field] = value;
        setFormState({ ...formState, FilesizePolicies: newPolicies });
    };

    const addFilePolicy = () => {
        setFormState({
            ...formState,
            FilesizePolicies: [...formState.FilesizePolicies, { Size: 1, Unit: 'KB', Percent: 0 }]
        });
    };

    const removeFilePolicy = (index) => {
        const newPolicies = formState.FilesizePolicies.filter((_, i) => i !== index);
        setFormState({ ...formState, FilesizePolicies: newPolicies });
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
            <DialogTitle>{campaign ? 'Edit Campaign' : 'New Campaign'}</DialogTitle>
            <DialogContent>
                <TextField
                    fullWidth
                    margin="normal"
                    label="Campaign Name"
                    name="name"
                    value={formState.name}
                    onChange={handleChange}
                />

                <FormControl fullWidth margin="normal">
                    <InputLabel>Protocol</InputLabel>
                    <Select
                        name="Protocol"
                        value={formState.Protocol}
                        onChange={handleChange}
                        label="Protocol"
                    >
                        <MenuItem value="FTP">FTP</MenuItem>
                        <MenuItem value="SFTP">SFTP</MenuItem>
                        <MenuItem value="HTTP">HTTP</MenuItem>
                        <MenuItem value="HTTPS">HTTPS</MenuItem>
                    </Select>
                </FormControl>

                <FormControl fullWidth margin="normal">
                    <InputLabel>Transfer Type</InputLabel>
                    <Select
                        name="Type"
                        value={formState.Type}
                        onChange={handleChange}
                        label="Transfer Type"
                    >
                        <MenuItem value="Download">Download</MenuItem>
                        <MenuItem value="Upload">Upload</MenuItem>
                    </Select>
                </FormControl>

                <Box sx={{ display: 'flex', gap: 2 }}>
                    <TextField
                        fullWidth
                        margin="normal"
                        label="Hostname"
                        name="Hostname"
                        value={formState.Hostname}
                        onChange={handleChange}
                    />
                    <TextField
                        margin="normal"
                        label="Port"
                        name="Port"
                        type="number"
                        value={formState.Port}
                        onChange={handleChange}
                        sx={{ width: 120 }}
                    />
                </Box>

                <TextField
                    fullWidth
                    margin="normal"
                    label="Path"
                    name="Path"
                    value={formState.Path}
                    onChange={handleChange}
                />

                <TextField
                    fullWidth
                    margin="normal"
                    label="Timeout (seconds)"
                    name="Timeout"
                    type="number"
                    value={formState.Timeout}
                    onChange={handleChange}
                />

                <Divider sx={{ my: 3 }} />
                <Typography variant="h6" gutterBottom>File Size Distribution</Typography>

                {formState.FilesizePolicies.map((policy, index) => (
                    <Box key={index} sx={{ display: 'flex', gap: 2, alignItems: 'center', mb: 2 }}>
                        <TextField
                            label="Size"
                            type="number"
                            value={policy.Size}
                            onChange={(e) => handleFilePolicyChange(index, 'Size', parseInt(e.target.value) || 0)}
                            sx={{ width: 100 }}
                        />
                        <FormControl sx={{ width: 120 }}>
                            <InputLabel>Unit</InputLabel>
                            <Select
                                value={policy.Unit}
                                onChange={(e) => handleFilePolicyChange(index, 'Unit', e.target.value)}
                                label="Unit"
                            >
                                <MenuItem value="KB">KB</MenuItem>
                                <MenuItem value="MB">MB</MenuItem>
                            </Select>
                        </FormControl>
                        <TextField
                            label="Percentage"
                            type="number"
                            value={policy.Percent}
                            onChange={(e) => handleFilePolicyChange(index, 'Percent', parseInt(e.target.value) || 0)}
                            sx={{ width: 120 }}
                        />
                        {formState.FilesizePolicies.length > 1 && (
                            <IconButton onClick={() => removeFilePolicy(index)}>
                                <RemoveIcon />
                            </IconButton>
                        )}
                        {index === formState.FilesizePolicies.length - 1 && (
                            <IconButton onClick={addFilePolicy}>
                                <AddIcon />
                            </IconButton>
                        )}
                    </Box>
                ))}
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button
                    onClick={() => onSave(formState)}
                    variant="contained"
                    color="primary"
                >
                    Save
                </Button>
            </DialogActions>
        </Dialog>
    );
} 