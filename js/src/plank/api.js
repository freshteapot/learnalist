import { copyObject } from '../utils/utils.js';
import { loggedIn, api } from "../shared.js";

async function today() {
    if (!loggedIn()) {
        return null;
    }

    const d = new Date();
    const datestring = `${d.getFullYear()}${d.getMonth() + 1}${d.getDate()}`;
    const labels = `plank.${datestring}`;

    try {
        let data = await api.getListsByMe({
            labels: labels
        });

        console.log("all my lists", data);
        if (data.length === 0) {
            // TODO create today
            try {
                data = await api.addList({
                    data: [],
                    info: {
                        title: "Todays planks",
                        type: "v1",
                        labels: [
                            "plank",
                            labels
                        ]
                    }
                });
                return convertFromV1(data);
            } catch (error) {
                console.error("yo2", error);
            }
        }
        return convertFromV1(data[0]);
    } catch (error) {
        console.log("yo");
        console.error(error);
        throw (error);
    }
}

// if I kept all,
// then it would be easier to delete
async function history() {

    try {
        let items = await api.getListsByMe({
            labels: "plank"
        });

        if (items.length === 0) {
            return [];
        }

        const reduced = items.reduce(function (filtered, item) {
            try {
                const copy = convertFromV1(item);
                filtered.push(copy);
            } catch (error) {

            }
            return filtered;
        }, []);
        return reduced;
    } catch (error) {
        console.error("history", error);
        throw (error);
    }
}

async function save(aList) {
    let input = convertToV1(aList);
    try {
        let aList = await api.updateList(input);
        console.log("planks", aList);
        return convertFromV1(aList);
    } catch (error) {
        console.error("yo2", error);
        throw (error);
    }
}

function convertToV1(aList) {
    let copy = copyObject(aList)
    copy.data = copy.data.map(e => {
        delete e.uuid;
        delete e.listIndex;
        return JSON.stringify(e);
    })
    return copy;
}

function convertFromV1(aList) {
    let copy = copyObject(aList)
    copy.data = copy.data.map((e, index) => {
        const obj = JSON.parse(e);
        obj.uuid = aList.uuid;
        obj.listIndex = index;
        return obj;
    })
    return copy;
}

export {
    today,
    history,
    save
}

