
import { getSpacedRepetitionNext, addSpacedRepetitionEntry, updateSpacedRepetitionEntry } from '../../../api2.js';
async function getNext() {
    return await getSpacedRepetitionNext();
}

async function viewed(uuid) {
    const input = {
        uuid: uuid,
        action: "incr"
    }
    return await updateSpacedRepetitionEntry(input)
}

async function addEntry(input) {
    return await addSpacedRepetitionEntry(input);
}

export {
    getNext,
    viewed,
    addEntry,
};
