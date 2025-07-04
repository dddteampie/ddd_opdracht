import callApi from './api';

export const getClients = async () => {
  return callApi('ecd', '/api/client');
};

export const addClient = async (clientData) => {
  return callApi('behoefte', '/ecd/client', { 
    method: 'POST',
    body: JSON.stringify(clientData),
  });
};
