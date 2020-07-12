
import { api } from '../../../store.js';
async function getNext() {
    return await api.getSpacedRepetitionNext();
}

async function viewed(uuid) {
    const input = {
        uuid: uuid,
        action: "incr"
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
