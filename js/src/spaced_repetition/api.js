
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

async function overtimeIsActive(uuid) {
    try {
        return await api.spacedRepetitionOvertimeIsActive(uuid);
    } catch (e) {
        console.log(e);
        return false;
    }
}

async function addListToOvertime(input) {
    return await api.spacedRepetitionAddListToOvertime(input);
}

async function removeListFromOvertime(userUuid, alistUuid) {
    return api.spacedRepetitionRemoveListFromOvertime(userUuid, alistUuid);
}

export {
    getEntries,
    getNext,
    viewed,
    addEntry,
    overtimeIsActive,
    addListToOvertime,
    removeListFromOvertime
};
