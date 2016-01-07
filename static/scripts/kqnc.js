function fixWrapperHeight() {
    // Get and set current height
    var headerH = 62;
    var navigationH = $("#navigation").height();
    var contentH = $(".content").height();

    // Set new height when contnet height is less then navigation
    if (contentH < navigationH) {
        $("#wrapper").css("min-height", navigationH + 'px');
    }

    // Set new height when contnet height is less then navigation and navigation is less then window
    if (contentH < navigationH && navigationH < $(window).height()) {
        $("#wrapper").css("min-height", $(window).height() - headerH  + 'px');
    }

    // Set new height when contnet is higher then navigation but less then window
    if (contentH > navigationH && contentH < $(window).height()) {
        $("#wrapper").css("min-height", $(window).height() - headerH + 'px');
    }
}

function setBodySmall() {
    if ($(this).width() < 769) {
        $('body').addClass('page-small');
    } else {
        $('body').removeClass('page-small');
    }
}

var App = function () {

    var currentPage = ''; 
    var collapsed = false;
    var is_mobile = false;
    var is_mini_menu = false;
    var is_fixed_header = false;
    var responsiveFunctions = [];

    var handleTooltips = function() {
        $('.tooltip-demo').tooltip({
            selector: "[data-toggle=tooltip]"
        });
    }

    var handlePopovers = function() {
        $("[data-toggle=popover]").popover();
    }

    var handleBodyModal = function() {
        $('.modal').appendTo("body");
    }

    var handleSparkline = function() {
        $("#sparkline1").sparkline([5, 6, 7, 2, 0, 4, 2, 4, 5, 7, 2, 4, 12, 11, 4], {
            type: 'bar',
            barWidth: 7,
            height: '30px',
            barColor: '#62cb31',
            negBarColor: '#53ac2a'
        });
    }

    var handleCheckboxes = function() {
        // Initialize iCheck plugin
        $('.i-checks').iCheck({
            checkboxClass: 'icheckbox_square-green',
            radioClass: 'iradio_square-green'
        });
    }

    var handleHeader = function() {
        // Function for small header
        $('.small-header-action').click(function(event){
            event.preventDefault();
            var icon = $(this).find('i:first');
            var breadcrumb  = $(this).parent().find('#hbreadcrumb');
            $(this).parent().parent().parent().toggleClass('small-header');
            breadcrumb.toggleClass('m-t-lg');
            icon.toggleClass('fa-arrow-up').toggleClass('fa-arrow-down');
        });
    }

    var handleSidebar = function() {
        if ($.cookie('hide_sidebar') === '1') {
            $("body").addClass("hide-sidebar");
        }
        if ($.cookie('hide_sidebar') === '0') {
            $("body").addClass("show-sidebar");
        }
        // Handle minimalize sidebar menu
        $('.hide-menu').click(function(event){
            event.preventDefault();
            if ($("body").hasClass("show-sidebar")) {
                $("body").removeClass("show-sidebar");
                $("body").addClass("hide-sidebar");
                $.cookie('hide_sidebar', '1');
            } else {
                $("body").addClass("show-sidebar");
                $("body").removeClass("hide-sidebar");
                $.cookie('hide_sidebar', '0');
            }
        });

        // Initialize metsiMenu plugin to sidebar menu
        $('#left-menu').metisMenu();

        $('#right-menu').metisMenu({ toggle: true });


        // Open close right sidebar
        $('.right-sidebar-toggle').click(function () {
            $('#right-sidebar').toggleClass('sidebar-open');
        });
    }

    var handlePanels = function() {
        // Function for collapse hpanel
        $('.showhide').click(function (event) {
            event.preventDefault();
            var hpanel = $(this).closest('div.hpanel');
            var icon = $(this).find('i:first');
            var body = hpanel.find('div.panel-body');
            var footer = hpanel.find('div.panel-footer');
            body.slideToggle(300);
            footer.slideToggle(200);

            // Toggle icon from up to down
            icon.toggleClass('fa-chevron-up').toggleClass('fa-chevron-down');
            hpanel.toggleClass('').toggleClass('panel-collapse');
            setTimeout(function () {
                hpanel.resize();
                hpanel.find('[id^=map-]').resize();
            }, 50);
        });

        // Function for close hpanel
        $('.closebox').click(function (event) {
            event.preventDefault();
            var hpanel = $(this).closest('div.hpanel');
            hpanel.remove();
        });

        // Fullscreen for fullscreen hpanel
        $('.fullscreen').click(function() {
            var hpanel = $(this).closest('div.hpanel');
            var icon = $(this).find('i:first');
            $('body').toggleClass('fullscreen-panel-mode');
            icon.toggleClass('fa-expand').toggleClass('fa-compress');
            hpanel.toggleClass('fullscreen');
            setTimeout(function() {
                $(window).trigger('resize');
            }, 100);
        });
    }

    var handleAnimate = function() {
        // Initialize animate panel function
        $('.animate-panel').animatePanel();
    }

    var handleDashboard = function() {
        var data1 = [ [0, 55], [1, 48], [2, 40], [3, 36], [4, 40], [5, 60], [6, 50], [7, 51] ];
        var data2 = [ [0, 56], [1, 49], [2, 41], [3, 38], [4, 46], [5, 67], [6, 57], [7, 59] ];

        var chartUsersOptions = {
            series: {
                splines: {
                    show: true,
                    tension: 0.4,
                    lineWidth: 1,
                    fill: 0.4
                },
            },
            grid: {
                tickColor: "#f0f0f0",
                borderWidth: 1,
                borderColor: 'f0f0f0',
                color: '#6a6c6f'
            },
            colors: [ "#62cb31", "#efefef"],
        };

        $.plot($("#flot-line-chart"), [data1, data2], chartUsersOptions);

        /**
         * Flot charts 2 data and options
         */
        var chartIncomeData = [
            {
                label: "line",
                data: [ [1, 10], [2, 26], [3, 16], [4, 36], [5, 32], [6, 51] ]
            }
        ];

        var chartIncomeOptions = {
            series: {
                lines: {
                    show: true,
                    lineWidth: 0,
                    fill: true,
                    fillColor: "#64cc34"

                }
            },
            colors: ["#62cb31"],
            grid: {
                show: false
            },
            legend: {
                show: false
            }
        };

        $.plot($("#flot-income-chart"), chartIncomeData, chartIncomeOptions);
    }

    var handleListingModals = function() {
        $("#PreviewModal iframe").load(function() {
          $(this).css("visibility", "visible");
        });
        $("#ListingsGallery a").click(function(e) {
            e.preventDefault();
            $("#PreviewModal").find(".modal-title").html($(this).data("title"));
            $("#PreviewModal").find("iframe").css('visibility', 'hidden');
            $("#PreviewModal").find("iframe").prop("src", $(this).prop("href"));
            $("#PreviewModal").modal("show");
        });
    }
    var handleDocumentView = function() {
        var textarea = document.getElementById("code1");

        // Wait until animation finished render container
        setTimeout(function(){

            CodeMirror.fromTextArea(textarea, {
                lineNumbers: true,
                matchBrackets: true,
                styleActiveLine: true
            });
        }, 500);
    }

    return {

        init: function () {
            setBodySmall();
            handleSidebar();
            handleCheckboxes();
            handleAnimate();
            handlePanels();
            handleHeader();
            // Set minimal height of #wrapper to fit the window
            setTimeout(function () {
                fixWrapperHeight();
            });
            handleSparkline();
            handleTooltips();
            handlePopovers();
            handleBodyModal();
            if (App.isPage("dashboard")) {
                handleDashboard();
            }
            if (App.isPage("document")) {
                handleDocumentView();
                handleListingModals();
            }
        },
        setPage: function (name) {
            currentPage = name;
        },
        isPage: function (name) {
            return currentPage == name ? true : false;
        },
        //public function to add callback a function which will be called on window resize
        addResponsiveFunction: function (func) {
            responsiveFunctions.push(func);
        },
        scrollTo: function (el, offeset) {
            pos = (el && el.size() > 0) ? el.offset().top : 0;
            jQuery('html,body').animate({
                scrollTop: pos + (offeset ? offeset : 0)
            }, 'slow');
        },
        scrollTop: function () {
            App.scrollTo();
        },
        // wrapper function to  block element(indicate loading)
        blockUI: function (el, loaderOnTop) {
            lastBlockedUI = el;
            jQuery(el).block({
                message: '<img src="./images/loading-bars.svg" align="absmiddle">',
                css: {
                    border: 'none',
                    padding: '2px',
                    backgroundColor: 'none'
                },
                overlayCSS: {
                    backgroundColor: '#000',
                    opacity: 0.05,
                    cursor: 'wait'
                }
            });
        },

        // wrapper function to  un-block element(finish loading)
        unblockUI: function (el) {
            jQuery(el).unblock({
                onUnblock: function () {
                    jQuery(el).removeAttr("style");
                }
            });
        },
    }
}();

$(window).bind("load", function () {
    // Remove splash screen after load
    $('.splash').css('display', 'none')
});

$(window).bind("resize click", function () {
    // Add special class to minimalize page elements when screen is less than 768px
    setBodySmall();
    // Waint until metsiMenu, collapse and other effect finish and set wrapper height
    setTimeout(function () {
        fixWrapperHeight();
    }, 300);
});

$.fn['animatePanel'] = function() {

    var element = $(this);
    var effect = $(this).data('effect');
    var delay = $(this).data('delay');
    var child = $(this).data('child');

    // Set default values for attrs
    if(!effect) { effect = 'zoomIn'}
    if(!delay) { delay = 0.06 } else { delay = delay / 10 }
    if(!child) { child = '.row > div'} else {child = "." + child}

    //Set defaul values for start animation and delay
    var startAnimation = 0;
    var start = Math.abs(delay) + startAnimation;

    // Get all visible element and set opacity to 0
    var panel = element.find(child);
    panel.addClass('opacity-0');

    // Get all elements and add effect class
    panel = element.find(child);
    panel.addClass('stagger').addClass('animated-panel').addClass(effect);

    var panelsCount = panel.length + 10;
    var animateTime = (panelsCount * delay * 10000) / 10;

    // Add delay for each child elements
    panel.each(function (i, elm) {
        start += delay;
        var rounded = Math.round(start * 10) / 10;
        $(elm).css('animation-delay', rounded + 's');
        // Remove opacity 0 after finish
        $(elm).removeClass('opacity-0');
    });

    // Clear animation after finish
    setTimeout(function(){
        $('.stagger').css('animation', '');
        $('.stagger').removeClass(effect).removeClass('animated-panel').removeClass('stagger');
    }, animateTime)

};