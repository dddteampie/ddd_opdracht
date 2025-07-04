import callApi from './api';

export const addBehoefte = async (behoefteData) => {
  return callApi("behoefte", '/behoefte', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(behoefteData),
  });
};

export const startCategorieAanvraag = async (inputData) => {
  return callApi("aanvraag", '/aanvraag/categorie', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(inputData),
  });
};

export const getPassendeCategorieenLijst = async (clientId) => {
  return callApi("aanvraag", `/aanvraag/recommendatie/categorie/?patientId=${clientId}`);
};

export const kiesCategorie = async (inputData) => {
  return callApi("aanvraag", '/aanvraag/categorie/kies', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(inputData),
  });
};

export const startProductAanvraag = async (inputData) => {
  return callApi("aanvraag", '/aanvraag/product', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(inputData),
  });
};

export const getPassendeProductenLijst = async (clientId) => {
  return callApi("aanvraag", `/aanvraag/recommendatie/product/?clientId=${clientId}`);
};

export const getZorgdossierByClientId = async (clientId) => {
    try {
        const response = await callApi("ecd", `/api/zorgdossier/client/${clientId}`);
        return response && response.id ? response : null;
    } catch (error) {
        console.error("Error fetching zorgdossier:", error);
        return null;
    }
};

export const getOnderzoekByDossierId = async (zorgdossierId) => {
    try {
        const response = await callApi("ecd", `/api/onderzoek/dossier/${zorgdossierId}`);
        return response || null;
    } catch (error) {
        console.error("Error fetching onderzoek by dossier ID:", error);
        return null;
    }
};

export const getOnderzoekByOnderzoekId = async (zorgdossierId) => {
    try {
        const response = await callApi("ecd", `/api/onderzoek/${zorgdossierId}`);
        return response && response.length > 0 ? response[0] : null;
    } catch (error) {
        console.error("Error fetching onderzoek by zorgdossier ID:", error);
        return null;
    }
};

export const createZorgdossier = async (zorgdossierData) => {
    return await callApi("behoefte", "/ecd/zorgdossier", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(zorgdossierData),
    });
};

export const createOnderzoek = async (onderzoekData) => {
    return await callApi("behoefte", "/ecd/onderzoek", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(onderzoekData),
    });
};

export const createDiagnose = async (onderzoekId, data) => {
  return callApi("behoefte", `/ecd/onderzoek/${onderzoekId}/diagnose`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
};

export const kiesProduct = async (inputData) => {
  return callApi("aanvraag", '/aanvraag/product/kies', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(inputData),
  });
};

export const getClient = async (clientId) => {
    try {
        const response = await callApi("ecd", `/api/client/${clientId}`);
        return response;
    } catch (error) {
        console.error(`Error fetching client with ID ${clientId}:`, error);
        throw error;
    }
};

export const getBehoeftenByClientId = async (clientId) => {
  return callApi("behoefte", `/behoefte/client/${clientId}`);
};

export const startAanvraag = async (aanvraagData, behoefteId) => {
  console.log(aanvraagData, behoefteId)
  return callApi("behoefte", `/behoefte/${behoefteId}/aanvraagverwerking`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(aanvraagData),
  });
};

export const getAanvragenByClientId = async (clientId) => {
  return callApi("aanvraag", `/aanvraag/client/${clientId}`);
};