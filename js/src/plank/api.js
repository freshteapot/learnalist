import { api } from "../shared.js";


async function deleteEntry(uuid) {
    try {
        return await api.deletePlankEntry(uuid);
    } catch (error) {
        console.error("history", error);
        throw (error);
    }
}

// if I kept all,
// then it would be easier to delete
async function history() {
    try {
        return await api.getPlankHistoryByUser();
    } catch (error) {
        console.error("history", error);
        throw (error);
    }
}

// TODO one by one
// VS all at once and return
async function saveEntry(entry) {
    try {
        return await api.addPlankEntry(entry);
    } catch (error) {
        console.error("yo2", error);
        throw (error);
    }
}

export {
    history,
    saveEntry,
    deleteEntry
}

