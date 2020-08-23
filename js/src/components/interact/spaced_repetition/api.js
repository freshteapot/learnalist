
import { api } from '../../../shared.js';
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
    getNext,
    viewed,
    addEntry,
};
