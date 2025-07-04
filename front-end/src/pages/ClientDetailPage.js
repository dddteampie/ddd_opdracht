import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useHistory } from 'react-router-dom';
import { Card, Spin, message, Typography, Divider, Form, Input, Button, DatePicker } from 'antd';
import Container from '../common/Container';
import {
    getClient,
    createOnderzoek,
    createZorgdossier,
    getZorgdossierByClientId,
    getOnderzoekByDossierId,
    createDiagnose,
} from '../services/behoefteService';
import moment from 'moment';

const { Title, Paragraph } = Typography;
const { TextArea } = Input;

const ClientDetailPage = () => {
    const { clientId } = useParams();
    const history = useHistory();
    const [client, setClient] = useState(null);
    const [zorgdossier, setZorgdossier] = useState(null);
    const [onderzoek, setOnderzoek] = useState(null);
    const [diagnose, setDiagnose] = useState(null); // Deze state wordt nu ALLEEN gevuld bij SUCCESVOLLE submit van diagnoseformulier
    const [loading, setLoading] = useState(true);
    const [zorgdossierForm] = Form.useForm();
    const [onderzoekForm] = Form.useForm();
    const [diagnoseForm] = Form.useForm();

    const BEHOEFTE_PAGINA_PATH = `/clients/${clientId}/behoeften`;

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const fetchedClient = await getClient(clientId);
            if (!fetchedClient) {
                message.error("Cliënt niet gevonden.");
                setClient(null);
                setLoading(false);
                return;
            }
            setClient(fetchedClient);

            let fetchedZorgdossier = await getZorgdossierByClientId(clientId);
            setZorgdossier(fetchedZorgdossier);

            let fetchedOnderzoek = null;
            // let fetchedDiagnose = null; // Deze is niet meer nodig hier, we halen diagnose niet op bij initieel laden

            if (fetchedZorgdossier && fetchedZorgdossier.id) {
                fetchedOnderzoek = await getOnderzoekByDossierId(fetchedZorgdossier.id);
                setOnderzoek(fetchedOnderzoek);

                // We halen hier GEEN bestaande diagnose op. De gebruiker moet altijd een diagnose invullen.
                // Als er ALLES bestaat (client, zorgdossier, onderzoek, EN we hebben een diagnose state),
                // dan kunnen we eventueel doorsturen.
                // Echter, volgens jouw laatste instructie,
                // sturen we direct door NA de succesvolle diagnose *submit*.
            } else {
                setOnderzoek(null);
                setDiagnose(null); // Zorg ervoor dat diagnose null is als zorgdossier ontbreekt
            }

            // BELANGRIJK: De automatische navigatie bij het laden van de pagina is verwijderd,
            // omdat de gebruiker altijd een diagnose *moet* invullen op deze pagina.
            // Navigatie gebeurt alleen na een succesvolle `handleDiagnoseSubmit`.

        } catch (error) {
            message.error("Fout bij het laden van cliëntdetails of gekoppelde dossiers.");
            console.error("Error fetching client details/dossiers:", error);
            setClient(null);
            setZorgdossier(null);
            setOnderzoek(null);
            setDiagnose(null);
        } finally {
            setLoading(false);
        }
    }, [clientId, history, BEHOEFTE_PAGINA_PATH]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const handleCreateZorgdossier = async (values) => {
        setLoading(true);
        try {
            const zorgdossierData = {
                client_id: clientId,
                situatie: values.situatie,
            };
            const newZorgdossier = await createZorgdossier(zorgdossierData);
            message.success("Zorgdossier succesvol aangemaakt!");
            setZorgdossier(newZorgdossier);
            zorgdossierForm.resetFields();
            await fetchData(); // Herhaal fetch om de volgende stap (Onderzoek) te controleren
        } catch (error) {
            message.error("Fout bij het aanmaken van het zorgdossier.");
            console.error("Error creating zorgdossier:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateOnderzoek = async (values) => {
        setLoading(true);
        try {
            if (!zorgdossier || !zorgdossier.id) {
                message.error("Kan geen onderzoek aanmaken: Zorgdossier ID ontbreekt.");
                return;
            }

            const vandaag = moment();
            const onderzoekData = {
                zorgdossier_id: zorgdossier.id,
                begin_datum: values.begin_datum.toISOString(),
                eind_datum: values.eind_datum ? values.eind_datum.toISOString() : vandaag.add(1, 'year').toISOString(),
            };
            const newOnderzoek = await createOnderzoek(onderzoekData);
            message.success("Onderzoek succesvol aangemaakt!");
            setOnderzoek(newOnderzoek);
            onderzoekForm.resetFields();
            await fetchData(); // Herhaal fetch om de volgende stap (Diagnose) te controleren
        } catch (error) {
            message.error("Fout bij het aanmaken van het onderzoek.");
            console.error("Error creating onderzoek:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleDiagnoseSubmit = async (values) => {
        setLoading(true);
        try {
            if (!onderzoek || !onderzoek.id) {
                message.error("Kan diagnose niet toevoegen: Onderzoek ID ontbreekt.");
                return;
            }

            const diagnoseData = {
                onderzoek_id: onderzoek.id,
                diagnosecode: values.diagnosecode,
                naam: values.naam,
                toelichting: values.toelichting,
                datum: moment().toISOString(),
                status: "Actief",
            };

            // BELANGRIJK: We gaan er vanuit dat createDiagnose altijd slaagt (HTTP 200 OK)
            // en de net aangemaakte diagnose teruggeeft.
            const newDiagnose = await createDiagnose(onderzoek.id, diagnoseData);
            message.success("Diagnose succesvol aangemaakt!");
            setDiagnose(newDiagnose); // Stel de state in met de response van de net aangemaakte diagnose

            diagnoseForm.resetFields(); // Reset het formulier na succesvolle indiening
            // TODO: Dit is het punt waar we nu direct navigeren, zonder expliciet de diagnose opnieuw op te halen.
            // Navigeer direct na het succesvol aanmaken van de diagnose
            history.push({
                pathname: BEHOEFTE_PAGINA_PATH,
                state: {
                    zorgdossierId: zorgdossier.id,
                    onderzoekId: onderzoek.id,
                    diagnoseId: newDiagnose.id // Geef de ID van de zojuist gemaakte diagnose mee
                },
            });

        } catch (error) {
            message.error("Fout bij het aanmaken van de diagnose.");
            console.error("Diagnose fout:", error);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <Container>
                <Spin size="large" tip="Controleer cliënt en dossiers..." />
            </Container>
        );
    }

    if (!client) {
        return (
            <Container>
                <p>Cliënt niet gevonden of een fout opgetreden.</p>
            </Container>
        );
    }

    return (
        <Container>
            <Card style={{ marginBottom: '20px' }}>
                <Title level={2}>Cliëntdetails: {client.naam}</Title>
                <Paragraph>ID: {client.id}</Paragraph>
                <Paragraph>Adres: {client.adres}</Paragraph>
                <Paragraph>Geboortedatum: {moment(client.geboortedatum).format('DD-MM-YYYY')}</Paragraph>
            </Card>

            {!zorgdossier && (
                <>
                    <Divider />
                    <Card title="Zorgdossier Aanmaken" style={{ marginTop: '20px' }}>
                        <Paragraph>Er is nog geen zorgdossier voor deze cliënt. Maak deze nu aan.</Paragraph>
                        <Form
                            form={zorgdossierForm}
                            layout="vertical"
                            onFinish={handleCreateZorgdossier}
                        >
                            <Form.Item
                                label="Situatie/Initiële Opmerking"
                                name="situatie"
                                rules={[{ required: true, message: 'Vul een korte situatieschets in!' }]}
                            >
                                <TextArea rows={3} placeholder="Beschrijf de initiële situatie van de cliënt." />
                            </Form.Item>
                            <Form.Item>
                                <Button type="primary" htmlType="submit" loading={loading} block>
                                    Zorgdossier Aanmaken
                                </Button>
                            </Form.Item>
                        </Form>
                    </Card>
                </>
            )}

            {zorgdossier && !onderzoek && (
                <>
                    <Divider />
                    <Card title="Onderzoek Aanmaken" style={{ marginTop: '20px' }}>
                        <Paragraph>Er is nog geen onderzoek gekoppeld aan dit zorgdossier. Maak deze nu aan.</Paragraph>
                        <Form
                            form={onderzoekForm}
                            layout="vertical"
                            onFinish={handleCreateOnderzoek}
                        >
                            <Form.Item
                                label="Begin Datum Onderzoek"
                                name="begin_datum"
                                initialValue={moment()}
                                rules={[{ required: true, message: 'Selecteer de begindatum van het onderzoek!' }]}
                            >
                                <DatePicker style={{ width: '100%' }} format="YYYY-MM-DD" />
                            </Form.Item>
                            <Form.Item
                                label="Eind Datum Onderzoek (optioneel)"
                                name="eind_datum"
                            >
                                <DatePicker style={{ width: '100%' }} format="YYYY-MM-DD" />
                            </Form.Item>
                            <Form.Item>
                                <Button type="primary" htmlType="submit" loading={loading} block>
                                    Onderzoek Aanmaken
                                </Button>
                            </Form.Item>
                        </Form>
                    </Card>
                </>
            )}

            {zorgdossier && onderzoek && ( // Altijd zichtbaar als zorgdossier en onderzoek bestaan
                <>
                    <Divider />
                    <Card title="Diagnose Toevoegen" style={{ marginTop: '20px' }}>
                        <Paragraph>Vul de diagnose in voor dit onderzoek. De ingevoerde diagnose wordt de actieve diagnose.</Paragraph>
                        <Form
                            form={diagnoseForm}
                            layout="vertical"
                            onFinish={handleDiagnoseSubmit}
                        >
                            <Form.Item
                                label="Diagnosecode"
                                name="diagnosecode"
                                rules={[{ required: true, message: 'Vul de diagnosecode in!' }]}
                            >
                                <Input placeholder="Bijv. G80.9 (Cerebrale Parese, ongespecificeerd)" />
                            </Form.Item>
                            <Form.Item
                                label="Naam"
                                name="naam"
                                rules={[{ required: true, message: 'Vul de naam van de diagnose in!' }]}
                            >
                                <Input placeholder="Bijv. Cerebrale Parese" />
                            </Form.Item>
                            <Form.Item
                                label="Toelichting"
                                name="toelichting"
                                rules={[{ required: true, message: 'Geef een toelichting op de diagnose!' }]}
                            >
                                <TextArea rows={3} placeholder="Gedetailleerde beschrijving van de diagnose..." />
                            </Form.Item>
                            <Form.Item>
                                <Button type="primary" htmlType="submit" loading={loading} block>
                                    Diagnose Toevoegen
                                </Button>
                            </Form.Item>
                        </Form>
                    </Card>
                </>
            )}

            {zorgdossier && (
                <>
                    <Divider />
                    <Card title="Bestaand Zorgdossier" style={{ marginTop: '20px' }}>
                        <Paragraph>ID: {zorgdossier.id}</Paragraph>
                        <Paragraph>Situatie: {zorgdossier.situatie}</Paragraph>
                    </Card>
                </>
            )}

            {onderzoek && (
                <>
                    <Divider />
                    <Card title="Bestaand Onderzoek" style={{ marginTop: '20px' }}>
                        <Paragraph>ID: {onderzoek.id}</Paragraph>
                        <Paragraph>Begindatum: {moment(onderzoek.begin_datum).format('DD-MM-YYYY')}</Paragraph>
                        <Paragraph>Einddatum: {moment(onderzoek.eind_datum).format('DD-MM-YYYY')}</Paragraph>
                    </Card>
                </>
            )}

            {diagnose && ( // Toon de zojuist aangemaakte diagnose als deze in de state staat
                <>
                    <Divider />
                    <Card title="Laatst Geregistreerde Diagnose" style={{ marginTop: '20px' }}>
                        <Paragraph>ID: {diagnose.id}</Paragraph>
                        <Paragraph>Onderzoek ID: {diagnose.onderzoek_id}</Paragraph>
                        <Paragraph>Code: {diagnose.diagnosecode}</Paragraph>
                        <Paragraph>Naam: {diagnose.naam}</Paragraph>
                        <Paragraph>Toelichting: {diagnose.toelichting}</Paragraph>
                        <Paragraph>Datum: {moment(diagnose.datum).format('DD-MM-YYYY HH:mm')}</Paragraph>
                        <Paragraph>Status: {diagnose.status}</Paragraph>
                    </Card>
                </>
            )}

            {client && zorgdossier && onderzoek && diagnose && ( // Navigatieknop alleen als alles compleet is
                 <>
                    <Divider />
                    <Button
                        type="primary"
                        onClick={() => history.push({
                            pathname: BEHOEFTE_PAGINA_PATH,
                            state: {
                                zorgdossierId: zorgdossier.id,
                                onderzoekId: onderzoek.id,
                                diagnoseId: diagnose.id
                            },
                        })}
                        style={{ marginTop: '20px' }}
                    >
                        Ga naar Behoeftenbepaling
                    </Button>
                 </>
            )}
        </Container>
    );
};

export default ClientDetailPage;