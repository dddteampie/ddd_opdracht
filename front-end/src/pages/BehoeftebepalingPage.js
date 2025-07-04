import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useHistory, useLocation } from 'react-router-dom';
import { Button, Input, List, Card, Spin, message, Form, Typography, Select } from 'antd'; // Select is toegevoegd voor urgentie
import Container from '../common/Container';
import {
    getBehoeftenByClientId, // Deze functie gaan we echt gebruiken
    addBehoefte,
    startCategorieAanvraag,
    getPassendeCategorieenLijst,
    kiesCategorie,
    getZorgdossierByClientId, // Gebruikt voor fallback om zorgdossier ID te vinden
    getOnderzoekByDossierId // Gebruikt voor fallback om onderzoek ID te vinden
} from '../services/behoefteService';

const { Title, Paragraph } = Typography;
const { Option } = Select; // Voor de urgentie dropdown

const BehoeftebepalingPage = () => {
    const { clientId } = useParams();
    const history = useHistory();
    const location = useLocation();

    const [zorgdossierId, setZorgdossierId] = useState(null);
    const [onderzoekId, setOnderzoekId] = useState(null);
    const [behoeften, setBehoeften] = useState([]);
    const [loading, setLoading] = useState(true);
    const [form] = Form.useForm();
    const [selectedBehoefte, setSelectedBehoefte] = useState(null);
    const [categorieen, setCategorieen] = useState([]);

    // Functie om de benodigde ID's te laden en behoeften op te halen
    const loadIdsAndFetchBehoeften = useCallback(async () => {
        setLoading(true);
        let currentZorgdossierId = null;
        let currentOnderzoekId = null;

        if (location.state && location.state.zorgdossierId && location.state.onderzoekId) {
            currentZorgdossierId = location.state.zorgdossierId;
            currentOnderzoekId = location.state.onderzoekId;
        } else {
            message.warning("Navigatiestate leeg. Probeer benodigde ID's via API op te halen...");
            try {
                const zorgdossier = await getZorgdossierByClientId(clientId);
                if (zorgdossier && zorgdossier.id) {
                    currentZorgdossierId = zorgdossier.id;
                    const onderzoek = await getOnderzoekByDossierId(currentZorgdossierId); // Haal onderzoek op via zorgdossier ID
                    if (onderzoek && onderzoek.id) {
                        currentOnderzoekId = onderzoek.id;
                    } else {
                        message.error("Geen gekoppeld onderzoek gevonden bij dit zorgdossier.");
                    }
                } else {
                    message.error("Geen zorgdossier gevonden voor deze cliënt.");
                }
            } catch (error) {
                message.error("Fout bij het ophalen van zorgdossier of gekoppeld onderzoek.");
                console.error("Error loading IDs for BehoeftebepalingPage:", error);
            }
        }

        setZorgdossierId(currentZorgdossierId);
        setOnderzoekId(currentOnderzoekId);

        if (currentOnderzoekId) {
            try {
                const fetchedBehoeften = await getBehoeftenByClientId(clientId); // Gebruik nu getBehoeftenByOnderzoekId
                setBehoeften(fetchedBehoeften);
                message.success("Behoeften succesvol geladen.");
            } catch (error) {
                message.error("Geen behoeften gevonden voor dit onderzoek.");
                console.error("Error fetching behoeften:", error);
                setBehoeften([]);
            }
        } else {
            message.info("Geen geldig Onderzoek ID beschikbaar om behoeften te laden.");
            setBehoeften([]);
        }
        setLoading(false);
    }, [clientId, location.state]);

    useEffect(() => {
        loadIdsAndFetchBehoeften();
    }, [loadIdsAndFetchBehoeften]);

    const handleAddBehoefte = async (values) => {
        const { titel, beschrijving, urgentie } = values;

        if (!titel.trim() || !beschrijving.trim()) {
            message.warning("Vul zowel de titel als de beschrijving in voor de behoefte.");
            return;
        }
        if (!onderzoekId) {
            message.error("Kan geen behoefte toevoegen: Onderzoek ID ontbreekt. Ververs de pagina of ga terug naar cliëntdetails.");
            return;
        }

        setLoading(true);
        try {
            const newBehoefteData = {
                onderzoek_id: onderzoekId, // Gebruik het opgehaalde onderzoekId (past bij backend payload)
                client_id: clientId,     // ClientId blijft via useParams beschikbaar (past bij backend payload)
                titel: titel,            // Nieuw: Titel
                beschrijving: beschrijving,
                urgentie: urgentie,      // Nieuw: Urgentie
                // status: 0 // Als status op backend wordt gegenereerd of optioneel is, kan deze weg
            };
            console.log(newBehoefteData)
            await addBehoefte(newBehoefteData);
            form.resetFields(); // Reset het formulier
            message.success("Behoefte succesvol toegevoegd.");
            await loadIdsAndFetchBehoeften(); // Herlaad behoeften na toevoegen
        } catch (error) {
            message.error("Fout bij het toevoegen van behoefte: " + error.message);
            console.error("Error adding behoefte:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleStartCategorieAanvraag = async (behoefte) => {
        setSelectedBehoefte(behoefte);
        setLoading(true);
        try {
            const inputData = {
                client_id: behoefte.client_id, // Gebruik de correcte veldnaam 'client_id'
                behoefte_beschrijving: behoefte.beschrijving
            };
            await startCategorieAanvraag(inputData);
            message.success("Categorie-aanvraag gestart. Categorieën ophalen...");

            const response = await getPassendeCategorieenLijst(behoefte.client_id); // Gebruik de correcte veldnaam 'client_id'
            setCategorieen(response.categorielijst);
            message.success("Categorieën succesvol opgehaald.");

        } catch (error) {
            message.error("Fout bij opvragen categorie aanbeveling: " + error.message);
            console.error("Error starting category request:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleKiesCategorie = async (categorieId) => {
        if (!selectedBehoefte) {
            message.error("Geen behoefte geselecteerd voor categoriekeuze.");
            return;
        }
        setLoading(true);
        try {
            const inputData = {
                client_id: selectedBehoefte.client_id, // Gebruik de correcte veldnaam 'client_id'
                behoefte_id: selectedBehoefte.id,
                categorie: categorieId
            };
            await kiesCategorie(inputData);
            message.success(`Categorie ${categorieId} gekozen.`);
            history.push(`/clients/${clientId}/aanvraag/${selectedBehoefte.id}/advies`);
        } catch (error) {
            message.error("Fout bij het kiezen van categorie: " + error.message);
            console.error("Error choosing category:", error);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <Container>
                <Spin size="large" tip="Behoeften laden..." />
            </Container>
        );
    }

    if (!zorgdossierId || !onderzoekId) {
        return (
            <Container>
                <Title level={3}>Fout: Zorgdossier of Onderzoek ID ontbreekt.</Title>
                <Paragraph>Controleer of de cliënt, het zorgdossier en het onderzoek correct zijn aangemaakt. Dit kan gebeuren na een paginaverversing als de benodigde gegevens niet direct via de URL beschikbaar zijn. Ga terug naar de cliëntdetails om dit te controleren.</Paragraph>
                <Button onClick={() => history.push(`/clients/${clientId}`)}>Terug naar Cliëntdetails</Button>
            </Container>
        );
    }

    return (
        <Container>
            <Title level={1}>Behoeften voor Cliënt: {clientId}</Title>
            <Paragraph>Huidig Zorgdossier ID: **{zorgdossierId}**</Paragraph>
            <Paragraph>Huidig Onderzoek ID: **{onderzoekId}**</Paragraph>

            <Card title="Nieuwe Behoefte Toevoegen" style={{ marginBottom: '20px' }}>
                <Form form={form} layout="vertical" onFinish={handleAddBehoefte}>
                    <Form.Item
                        label="Titel Behoefte"
                        name="titel"
                        rules={[{ required: true, message: 'Vul een titel in voor de behoefte!' }]}
                    >
                        <Input placeholder="Bijv. Lopen, Eten, Sociale Interactie" />
                    </Form.Item>
                    <Form.Item
                        label="Beschrijving Behoefte"
                        name="beschrijving"
                        rules={[{ required: true, message: 'Vul een beschrijving in voor de behoefte!' }]}
                    >
                        <Input.TextArea
                            rows={4}
                            placeholder="Bijv. Patiënt wil leren lopen na val"
                        />
                    </Form.Item>
                    <Form.Item
                        label="Urgentie"
                        name="urgentie"
                        initialValue="Laag" // Standaardwaarde
                        rules={[{ required: true, message: 'Selecteer de urgentie!' }]}
                    >
                        <Select placeholder="Selecteer urgentie">
                            <Option value="Laag">Laag</Option>
                            <Option value="Normaal">Normaal</Option>
                            <Option value="Hoog">Hoog</Option>
                        </Select>
                    </Form.Item>
                    <Button type="primary" htmlType="submit" loading={loading}>
                        Behoefte Toevoegen
                    </Button>
                </Form>
            </Card>

            <Card title="Bestaande Behoeften">
                <List
                    bordered
                    dataSource={behoeften}
                    renderItem={(behoefte) => (
                        <List.Item
                            actions={[
                                <Button key="advies" onClick={() => handleStartCategorieAanvraag(behoefte)}>
                                    Vraag Categorie Advies
                                </Button>,
                            ]}
                        >
                            <List.Item.Meta
                                title={behoefte.titel ? behoefte.titel : behoefte.beschrijving}
                                description={
                                    <>
                                        <Paragraph>Beschrijving: {behoefte.beschrijving}</Paragraph>
                                        <Paragraph>Urgentie: {behoefte.urgentie || 'Niet gespecificeerd'}</Paragraph>
                                        <Paragraph>Status: {behoefte.status || 'Niet gespecificeerd'} | Datum: {new Date(behoefte.datum).toLocaleDateString()}</Paragraph>
                                    </>
                                }
                            />
                        </List.Item>
                    )}
                />
            </Card>

            {categorieen.length > 0 && (
                <Card title="Gevonden Categorieën" style={{ marginTop: '20px' }}>
                    <List
                        bordered
                        dataSource={categorieen}
                        renderItem={(categorie) => (
                            <List.Item
                                actions={[
                                    <Button key="select-cat" onClick={() => handleKiesCategorie(categorie.ID)}>
                                        Kies deze Categorie
                                    </Button>,
                                ]}
                            >
                                <List.Item.Meta
                                    title={categorie.Naam}
                                    description={`ID: ${categorie.ID}`}
                                />
                            </List.Item>
                        )}
                    />
                </Card>
            )}
        </Container>
    );
};

export default BehoeftebepalingPage;