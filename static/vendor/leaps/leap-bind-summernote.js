/*
Copyright (c) 2014 Ashley Jeffs

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, sub to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

/*jshint newcap: false*/

(function() {
"use strict";

/*--------------------------------------------------------------------------------------------------
 */

/* leap_bind_summernote takes an existing leap_client and uses it to wrap a summernote into an
 * interactive editor for the leaps document the client connects to. Returns the bound object, and
 * places any errors in the obj.error field to be checked after construction.
 */
var leap_bind_summernote = function(leap_client, summernote) {
	this._summernote = summernote;
	this._leap_client = leap_client;

	this._content = "";
	this._ready = false;
	this._summernote.summernote('disable');

	var binder = this;


	$('.summernote').on('summernote.keyup', function(we, e) {
	  binder._trigger_diff();
	});

	/*summernote.summernote({onChange: function() {
		binder._trigger_diff(); }});*/

	this._leap_client.subscribe_event("document", function(doc) {
		console.log(doc.content);
		//binder._content = binder._summernote.value = doc.content;
		binder._summernote.summernote("code", doc.content);
		binder._content = doc.content;

		binder._ready = true;
		binder._summernote.summernote('enable');

		binder._pos_interval = setInterval(function() {
			binder._leap_client.update_cursor.apply(binder._leap_client, [ binder._summernote.selectionStart ]);
		}, leap_client._POSITION_POLL_PERIOD);
	});

	this._leap_client.subscribe_event("transforms", function(transforms) {
		for ( var i = 0, l = transforms.length; i < l; i++ ) {
			binder._apply_transform.apply(binder, [ transforms[i] ]);
		}
	});

	this._leap_client.subscribe_event("disconnect", function() {
		binder._summernote.summernote('disable');
		if ( undefined !== binder._pos_interval ) {
			clearTimeout(binder._pos_interval);
		}
	});

	this._leap_client.subscribe_event("user", function(user) {
		console.log("User update: " + JSON.stringify(user));
	});
};

/* apply_transform, applies a single transform to the summernote. Also attempts to retain the original
 * cursor position.
 */
leap_bind_summernote.prototype._apply_transform = function(transform) {
	var cursor_pos = this._summernote.selectionStart;
	var cursor_pos_end = this._summernote.selectionEnd;
	var content = this._summernote.summernote('code');

	if ( transform.position <= cursor_pos ) {
		cursor_pos += (transform.insert.length - transform.num_delete);
		cursor_pos_end += (transform.insert.length - transform.num_delete);
	}

	//this._content = this._summernote.value = this._leap_client.apply(transform, content);
	this._summernote.summernote('code', this._leap_client.apply(transform, content));
	this._content = this._leap_client.apply(transform, content);

	this._summernote.selectionStart = cursor_pos;
	this._summernote.selectionEnd = cursor_pos_end;
};

/* trigger_diff triggers whenever a change may have occurred to the wrapped summernote element, and
 * compares the old content with the new content. If a change has indeed occurred then a transform
 * is generated from the comparison and dispatched via the leap_client.
 */
leap_bind_summernote.prototype._trigger_diff = function() {
	var new_content = this._summernote.summernote('code');
	console.log("New:", new_content);
	console.log("cont:", this._content);
	if ( !(this._ready) || new_content === this._content ) {
		return;
	}

	var i = 0, j = 0;
	while (new_content[i] === this._content[i]) {
		i++;
	}
	while ((new_content[(new_content.length - 1 - j)] === this._content[(this._content.length - 1 - j)]) &&
			((i + j) < new_content.length) && ((i + j) < this._content.length)) {
		j++;
	}

	var tform = { position : i };

	if (this._content.length !== (i + j)) {
		tform.num_delete = (this._content.length - (i + j));
	}
	if (new_content.length !== (i + j)) {
		tform.insert = new_content.slice(i, new_content.length - j);
	}

	this._content = new_content;
	if ( tform.insert !== undefined || tform.num_delete !== undefined ) {
		var err = this._leap_client.send_transform(tform);
		if ( err !== undefined ) {
			this._leap_client._dispatch_event.apply(this._leap_client,
				[ this._leap_client.EVENT_TYPE.ERROR, [
					"Local change resulted in invalid transform"
				] ]);
		}
	}
};

/*--------------------------------------------------------------------------------------------------
 */

try {
	if ( window.leap_client !== undefined && typeof(window.leap_client) === "function" ) {
		window.leap_client.prototype.bind_summernote = function(summernote) {
			this._summernote = new leap_bind_summernote(this, summernote);
		};
	}
} catch (e) {
}

/*--------------------------------------------------------------------------------------------------
 */

})();