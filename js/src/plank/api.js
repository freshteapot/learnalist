import { copyObject } from '../utils/utils.js';
import { loggedIn } from "../store.js";
import { getListsByMe, addList, updateList } from "../api2.js";

async function today() {
    if (!loggedIn()) {
        return null;
    }

    const d = new Date();
    const datestring = `${d.getFullYear()}${d.getMonth() + 1}${d.getDate()}`;
    const labels = `plank.${datestring}`;
    console.log("labels", labels);

    try {
        let data = await getListsByMe({
            labels: labels
        });

        console.log("all my lists", data);
        if (data.length === 0) {
            // TODO create today
            try {
                data = await addList({
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
                console.log("planks", data);
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

async function history() {

    try {
        let items = await getListsByMe({
            labels: "plank"
        });

        if (items.length === 0) {
            return [];
        }

        const reduced = items.reduce(function (filtered, item) {
            try {
                const copy = convertFromV1(item);
                filtered.push(...copy.data);
            } catch (error) {

            }
            return filtered;
        }, []);
        return reduced;
        /*
        return [].concat(...
            items.filter(item => {
                try {
                    convertFromV1(item);
                    return true;
                } catch (error) {
                    return false;
                }
            }).map(item => {
                const copy = convertFromV1(item);
                return copy.data;
            }))
        */

    } catch (error) {
        console.error("history", error);
        throw (error);
    }
}

async function save(aList) {
    let input = convertToV1(aList);
    try {
        let aList = await updateList(input);
        console.log("planks", aList);
        return convertFromV1(aList);
    } catch (error) {
        console.error("yo2", error);
        throw (error);
    }
}

function convertToV1(aList) {
    let copy = copyObject(aList)
    copy.data = copy.data.map(e => JSON.stringify(e))
    return copy;
}

function convertFromV1(aList) {
    let copy = copyObject(aList)
    copy.data = copy.data.map(e => JSON.parse(e))
    return copy;
}

export {
    today,
    history,
    save
}

