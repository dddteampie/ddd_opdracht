import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useHistory, useLocation } from 'react-router-dom';
import { Button, Input, List, Card, Spin, message, Form, Typography, Select, Divider } from 'antd';
import Container from '../common/Container';
import {
    getBehoeftenByClientId,
    addBehoefte,
    startCategorieAanvraag,
    getPassendeCategorieenLijst,
    kiesCategorie,
    getZorgdossierByClientId,
    getOnderzoekByDossierId,
    // New functions for aanvraag and product flows
    startAanvraag,
    getAanvragenByClientId as fetchAanvragenByClientId, // Renamed to avoid conflict
    startProductAanvraag,
    getPassendeProductenLijst,
    kiesProduct,
} from '../services/behoefteService'; // Assuming new functions are added here

const { Title, Paragraph } = Typography;
const { Option } = Select;

const BehoeftebepalingPage = () => {
    const { clientId } = useParams();
    const history = useHistory();
    const location = useLocation();

    const [zorgdossierId, setZorgdossierId] = useState(null);
    const [onderzoekId, setOnderzoekId] = useState(null);
    const [behoeften, setBehoeften] = useState([]);
    const [aanvragen, setAanvragen] = useState([]); // New state for aanvragen
    const [loading, setLoading] = useState(true);
    const [form] = Form.useForm();
    const [selectedAanvraag, setSelectedAanvraag] = useState(null); // New state for selected aanvraag
    const [categorieen, setCategorieen] = useState([]);
    const [producten, setProducten] = useState([]); // New state for recommended products

    // Function to load necessary IDs and fetch behoeften and aanvragen
    const loadData = useCallback(async () => {
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
                    const onderzoek = await getOnderzoekByDossierId(currentZorgdossierId);
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
                const fetchedBehoeften = await getBehoeftenByClientId(clientId);
                setBehoeften(fetchedBehoeften);
                message.success("Behoeften succesvol geladen.");
            } catch (error) {
                message.info("Geen behoeften gevonden voor dit onderzoek.");
                console.error("Error fetching behoeften:", error);
                setBehoeften([]);
            }

            try {
                const fetchedAanvragen = await fetchAanvragenByClientId(clientId); // Use the new fetch
                setAanvragen(fetchedAanvragen);
                message.success("Aanvragen succesvol geladen.");
            } catch (error) {
                message.info("Geen aanvragen gevonden voor deze cliënt.");
                console.error("Error fetching aanvragen:", error);
                setAanvragen([]);
            }
        } else {
            message.info("Geen geldig Onderzoek ID beschikbaar om behoeften en aanvragen te laden.");
            setBehoeften([]);
            setAanvragen([]);
        }
        setLoading(false);
    }, [clientId, location.state]);

    useEffect(() => {
        loadData();
    }, [loadData]);

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
                onderzoek_id: onderzoekId,
                client_id: clientId,
                titel: titel,
                beschrijving: beschrijving,
                urgentie: urgentie,
            };
            console.log("Adding behoefte with data:", newBehoefteData);
            await addBehoefte(newBehoefteData);
            form.resetFields();
            message.success("Behoefte succesvol toegevoegd.");
            await loadData(); // Reload all data
        } catch (error) {
            message.error("Fout bij het toevoegen van behoefte: " + (error.message || error));
            console.error("Error adding behoefte:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleStartAanvraag = async (behoefte) => {
        setLoading(true);
        try {
            // Need client data for StartAanvraag handler. Assuming `behoefte` object has `client_id`.
            // In a real app, you might fetch client details or pass them from a parent context.
            const clientData = { id: clientId }; // Simplified client object for the payload
            const aanvraagPayload = {
                client: clientData,
                behoefte: behoefte,
            };
            console.log(aanvraagPayload, behoefte.id);
            const newAanvraag = await startAanvraag(aanvraagPayload, behoefte.id);
            message.success(`Aanvraag voor behoefte "${behoefte.titel}" succesvol gestart.`);
            setSelectedAanvraag(newAanvraag); // Set the newly created aanvraag as selected
            await loadData(); // Reload aanvragen list
        } catch (error) {
            message.error("Fout bij het starten van de aanvraag: " + (error.message || error));
            console.error("Error starting aanvraag:", error);
        } finally {
            setLoading(false);
        }
    };

const handleStartCategorieAanvraag = async (aanvraagToProcess) => {
    setSelectedAanvraag(aanvraagToProcess);
    setLoading(true);

    // Add a check to ensure `behoefte` exists on the `aanvraagToProcess` object
    if (!aanvraagToProcess.Behoefte || !aanvraagToProcess.Behoefte.beschrijving) {
        message.error("Beschrijving van de behoefte ontbreekt voor deze aanvraag. Kan geen categorie advies starten.");
        setLoading(false);
        return;
    }

    try {
        console.log(aanvraagToProcess.Behoefte)
        const inputData = {
            patientId: aanvraagToProcess.client_id,
            budget: aanvraagToProcess.budget || 20000,
            behoeften: aanvraagToProcess.Behoefte.beschrijving,
        };
        console.log(inputData)
        await startCategorieAanvraag(inputData);
        // Only show success if no error is thrown
        message.success("Categorie-aanvraag gestart. Categorieën ophalen...");

        const response = await getPassendeCategorieenLijst(aanvraagToProcess.client_id);
        if (response?.categories) {
            setCategorieen(response.categories);
            message.success("Categorieën succesvol opgehaald.");
        } else {
            message.warning("Geen categorieën gevonden in het antwoord.");
            setCategorieen([]);
        }
        await loadData();
    } catch (error) {
        message.error("Fout bij opvragen categorie aanbeveling: " + (error.message || error));
        console.error("Error starting category request:", error);
    } finally {
        setLoading(false);
    }
};

    const handleKiesCategorie = async (categorieId) => {
        if (!selectedAanvraag) {
            message.error("Geen aanvraag geselecteerd voor categoriekeuze.");
            return;
        }
        setLoading(true);
        try {
            const inputData = {
                client_id: selectedAanvraag.client_id,
                behoefte_id: selectedAanvraag.behoefte_id,
                categorie: categorieId,
            };
            await kiesCategorie(inputData);
            message.success(`Categorie ${categorieId} gekozen voor aanvraag ${selectedAanvraag.id}.`);
            await loadData(); // Refresh aanvragen to show updated status
            // Optionally, clear categories and prepare for product recommendation
            setCategorieen([]);
            setSelectedAanvraag({ ...selectedAanvraag, GekozenCategorieID: categorieId }); // Update selected Aanvraag locally
        } catch (error) {
            message.error("Fout bij het kiezen van categorie: " + (error.message || error));
            console.error("Error choosing category:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleStartProductAanvraag = async (aanvraagToProcess) => {
        setSelectedAanvraag(aanvraagToProcess); // Ensure the correct aanvraag is selected
        setLoading(true);
        try {
            if (!aanvraagToProcess.gekozen_categorie_id) {
                message.warning("Er is nog geen categorie gekozen voor deze aanvraag.");
                setLoading(false);
                return;
            }

            const inputData = {
                clientId: aanvraagToProcess.client_id,
                budget: aanvraagToProcess.budget || 20000,
                behoeften: aanvraagToProcess.Behoefte.beschrijving,
                CategorieID: aanvraagToProcess.gekozen_categorie_id,
            };
            console.log(JSON.stringify(inputData));
            await startProductAanvraag(inputData);
            message.success("Product-aanvraag gestart. Producten ophalen...");

            const response = await getPassendeProductenLijst(aanvraagToProcess.client_id);
            setProducten(response.products);
            message.success("Producten succesvol opgehaald.");
            await loadData(); // Refresh aanvragen to show updated status
        } catch (error) {
            message.error("Fout bij opvragen product aanbeveling: " + (error.message || error));
            console.error("Error starting product request:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleKiesProduct = async (productEAN) => {
        if (!selectedAanvraag) {
            message.error("Geen aanvraag geselecteerd voor productkeuze.");
            return;
        }
        setLoading(true);
        try {
            const inputData = {
                client_id: selectedAanvraag.client_id,
                behoefte_id: selectedAanvraag.behoefte_id,
                product_ean: productEAN,
            };
            await kiesProduct(inputData);
            message.success(`Product met EAN ${productEAN} gekozen voor aanvraag ${selectedAanvraag.id}.`);
            await loadData(); // Refresh aanvragen to show updated status
            setProducten([]); // Clear product list
            setSelectedAanvraag(null); // Clear selected aanvraag as workflow is complete
        } catch (error) {
            message.error("Fout bij het kiezen van product: " + (error.message || error));
            console.error("Error choosing product:", error);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <Container>
                <Spin size="large" tip="Gegevens laden..." />
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
            <Title level={1}>Behoeften en Aanvragen voor Cliënt: {clientId}</Title>
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
                        initialValue="Laag"
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

            <Divider />

            <Card title="Bestaande Behoeften" style={{ marginBottom: '20px' }}>
                <List
                    bordered
                    dataSource={behoeften}
                    renderItem={(behoefte) => (
                        <List.Item
                            actions={[
                                <Button
                                    key="start-aanvraag"
                                    onClick={() => handleStartAanvraag(behoefte)}
                                    disabled={loading || aanvragen.some(a => a.behoefte_id === behoefte.id)}
                                >
                                    Start Aanvraag
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

            <Divider />

            <Card title="Bestaande Aanvragen" style={{ marginBottom: '20px' }}>
                <List
                    bordered
                    dataSource={aanvragen}
                    renderItem={(aanvraag) => (
                        <List.Item
                            actions={[
                                // Only show "Vraag Categorie Advies" if no category is chosen yet and it's not waiting for a category choice
                                (!aanvraag.gekozen_categorie_id && aanvraag.status !== 1) && (
                                    <Button
                                        key="advies"
                                        onClick={() => handleStartCategorieAanvraag(aanvraag)}
                                        loading={loading && selectedAanvraag?.id === aanvraag.id}
                                    >
                                        Vraag Categorie Advies
                                    </Button>
                                ),
                                // Show "Vraag Product Advies" if a category is chosen and no product is chosen yet
                                (aanvraag.gekozen_categorie_id && !aanvraag.gekozen_product_id && aanvraag.status !== 2) && (
                                    <Button
                                        key="product-advies"
                                        onClick={() => handleStartProductAanvraag(aanvraag)}
                                        loading={loading && selectedAanvraag?.id === aanvraag.id}
                                    >
                                        Vraag Product Advies
                                    </Button>
                                ),
                            ]}
                        >
                            <List.Item.Meta
                                title={`Aanvraag voor: ${aanvraag.Behoefte?.titel || aanvraag.Behoefte?.beschrijving || 'Onbekende Behoefte'}`}
                                description={
                                    <>
                                        <Paragraph>Aanvraag ID: {aanvraag.id}</Paragraph>
                                        <Paragraph>Behoefte ID: {aanvraag.behoefte_id}</Paragraph>
                                        <Paragraph>Status: **{aanvraag.StatusString || aanvraag.status}**</Paragraph>
                                        <Paragraph>Budget: €{aanvraag.budget ? aanvraag.budget.toFixed(2) : 'N.v.t.'}</Paragraph>
                                        {aanvraag.gekozen_categorie_id && <Paragraph>Gekozen Categorie ID: {aanvraag.gekozen_categorie_id}</Paragraph>}
                                        {aanvraag.gekozen_product_id && <Paragraph>Gekozen Product EAN: {aanvraag.gekozen_product_id}</Paragraph>}
                                    </>
                                }
                            />
                        </List.Item>
                    )}
                />
            </Card>

            {categorieen.length > 0 && selectedAanvraag && (
                <Card title={`Gevonden Categorieën voor Aanvraag: ${selectedAanvraag.id}`} style={{ marginTop: '20px' }}>
                    <List
                        bordered
                        dataSource={categorieen}
                        renderItem={(categorie) => (
                            <List.Item
                                actions={[
                                    <Button key="select-cat" onClick={() => handleKiesCategorie(categorie.id)}>
                                        Kies deze Categorie
                                    </Button>,
                                ]}
                            >
                                <List.Item.Meta
                                    title={categorie.naam}
                                    description={`ID: ${categorie.id}`}
                                />
                            </List.Item>
                        )}
                    />
                </Card>
            )}

            {producten.length > 0 && selectedAanvraag?.gekozen_categorie_id && (
                <Card title={`Gevonden Producten voor Aanvraag: ${selectedAanvraag.id} (Categorie: ${selectedAanvraag.gekozen_categorie_id})`} style={{ marginTop: '20px' }}>
                    <List
                        bordered
                        dataSource={producten}
                        renderItem={(product) => (
                            <List.Item
                                actions={[
                                    <Button key="select-product" onClick={() => handleKiesProduct(product.ean)}>
                                        Kies dit Product
                                    </Button>,
                                ]}
                            >
                                <List.Item.Meta
                                    title={product.naam}
                                    description={
                                        <>
                                            <Paragraph>EAN: {product.ean}</Paragraph>
                                            <Paragraph>Prijs: €{product.prijs ? product.prijs.toFixed(2) : 'N.v.t.'}</Paragraph>
                                            <Paragraph>Beschrijving: {product.beschrijving}</Paragraph>
                                        </>
                                    }
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