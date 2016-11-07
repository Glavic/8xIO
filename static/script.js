var Binder = {
	objs: [],

	add: function(obj) {
		Binder.objs[Binder.objs.length] = obj;
	},

	run: function() {
		Binder.eval.run();
	},

	eval: {
		run: function() {
			$.each(Binder.objs, function(i, obj) {
				try {
					obj.run();
				} catch (err) {};
			});
		}
	}
}
function run(obj) {
	Binder.add(obj);
}
jQuery(function($) {
	Binder.run();
});


var Home = {
	obj: null,
	ajax_active: false,

	run: function() {
		Home.obj = $('#Home');
		if (!Home.obj.length) {
			return;
		}
		Home.obj.find('li').click(Home.changeState);
		setInterval(Home.checkState, 50);
	},

	changeState: function(e) {
		e.preventDefault();
		var li = $(this);
		$.get(li.attr('data-change-uri'));
	},

	checkState: function() {
		if (Home.ajax_active) {
			return;
		}
		Home.ajax_active = true;

		Home.ajax = $.ajax({
			url: '/state?w8',
			dataType: 'json',
			success: function(data, textStatus, jqXHR) {
				$.each(data, function(k, v) {
					for (bit = 0; bit < 8; bit++) {
						var state = v.out_bin[bit] == "1" ? "ON" : "OFF";
						Home.obj.find('li[data-addr="'+v.addr+'"][data-bit="'+(7-bit)+'"]').attr('data-state', state);
					}
				})
			},
			complete: function() {
				Home.ajax_active = false;
			}
		});
	}
}
run(Home);


var Config = {
	obj: null,

	run: function() {
		Config.obj = $('#Config');
		if (!Config.obj.length) {
			return;
		}
		Config.obj.find('td.state a').click(Config.changeState);
		Config.obj.find('td.name input').change(Config.changeName);
		Config.obj.find('td.ord input').change(Config.changeOrd);
	},

	changeState: function(e) {
		e.preventDefault();
		var link = $(this);
		var tr = link.closest('tr')
		$.get(link.attr('href'), function(data) {
			if (data == "ON" || data == "OFF") {
				tr.attr('data-state', data);
			} else {
				tr.attr('data-state', '');
				console.log("Response is invalid!");
			}
		});
	},

	changeName: function(e) {
		e.preventDefault();
		var input = $(this);
		input.removeClass('set')
		var tr = input.closest('tr')
		$.get(tr.find('td.state a').attr('href') + '&name=' + input.val(), function(data) {
			if (data == "NAME-SET") {
				input.addClass('set');
			}
		});
	},

	changeOrd: function(e) {
		e.preventDefault();
		var input = $(this);
		input.removeClass('set')
		var tr = input.closest('tr')
		$.get(tr.find('td.state a').attr('href') + '&ord=' + input.val(), function(data) {
			if (data == "ORD-SET") {
				input.addClass('set');
			}
		});
	}

}
run(Config);
