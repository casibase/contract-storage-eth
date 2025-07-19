// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

contract SaveContract {
    // Define struct
    struct DataItem {
        string key;
        string field;
        string value;
    }
    
    DataItem public data;

    // Define event, returns key, field, value in order
    event DataSaved(string key, string field, string value);

    // Save struct data and return via event
    function save(DataItem memory _data) public {
        data = _data;
        emit DataSaved(_data.key, _data.field, _data.value);
    }
    
    // Overloaded function: also supports passing three string parameters directly
    function save(string memory _key, string memory _field, string memory _value) public {
        data = DataItem(_key, _field, _value);
        emit DataSaved(_key, _field, _value);
    }
}