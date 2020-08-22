
import { api } from '../shared.js';

async function getEntries() {
    return await api.getSpacedRepetitionEntries();
}
async function getNext() {
    return await api.getSpacedRepetitionNext();
}

async function viewed(uuid, action) {
    const input = {
        uuid,
        action
    }
    return await api.updateSpacedRepetitionEntry(input)
}

async function addEntry(input) {
    return await api.addSpacedRepetitionEntry(input);
}

export {
    getEntries,
    getNext,
    viewed,
    addEntry,
};
