import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api',
  withCredentials: true
});

export const getTestHistory = (page = 1, size = 10) => 
  api.get(`/tests?page=${page}&size=${size}`);

export const getTestDetails = (testId) =>
  api.get(`/tests/${testId}`); 