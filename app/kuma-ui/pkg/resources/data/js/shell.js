(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["shell"],{"0b00":function(e,t,n){"use strict";n("7676")},"1b4f":function(e,t,n){},2467:function(e,t,n){"use strict";n("6dd1")},3372:function(e,t,n){"use strict";n("d846")},"42e7":function(e,t,n){"use strict";n("45cc")},"45cc":function(e,t,n){},"55b3":function(e,t,n){"use strict";n("1b4f")},"5f76":function(e,t,n){},"66b4":function(e,t,n){},"6dd1":function(e,t,n){},7148:function(e,t,n){"use strict";n("66b4")},"75bb":function(e,t,n){"use strict";n.d(t,"a",(function(){return a}));var a={PAGINATION_PREVIOUS_BUTTON_CLICKED:"pagination-previous-button-clicked",PAGINATION_NEXT_BUTTON_CLICKED:"pagination-next-button-clicked",SIDEBAR_ITEM_CLICKED:"sidebar-item-clicked",TABLE_REFRESH_BUTTON_CLICKED:"table-refresh-button-clicked",TABS_TAB_CHANGE:"tabs-tab-change",CREATE_MESH_CLICKED:"create-mesh-clicked",CREATE_DATA_PLANE_PROXY_CLICKED:"create-data-plane-proxy-clicked"}},7676:function(e,t,n){},"857a":function(e,t,n){var a=n("1d80"),s=/"/g;e.exports=function(e,t,n,i){var r=String(a(e)),o="<"+t;return""!==n&&(o+=" "+n+'="'+String(i).replace(s,"&quot;")+'"'),o+">"+r+"</"+t+">"}},9911:function(e,t,n){"use strict";var a=n("23e7"),s=n("857a"),i=n("af03");a({target:"String",proto:!0,forced:i("link")},{link:function(e){return s(this,"a","href",e)}})},af03:function(e,t,n){var a=n("d039");e.exports=function(e){return a((function(){var t=""[e]('"');return t!==t.toLowerCase()||t.split('"').length>3}))}},d846:function(e,t,n){},deb3:function(e,t,n){"use strict";n.r(t);var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"main-content-container"},[n("Sidebar"),n("main",{staticClass:"main-content"},[n("div",{staticClass:"page"},[e.showOnboarding?n("OnboardingCheck"):e._e(),n("Breadcrumbs"),n("router-view")],1)])],1)},s=[],i=n("f3f3"),r=n("2f62"),o=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("aside",{staticClass:"has-subnav",class:[{"is-collapsed":e.isCollapsed},{"subnav-expanded":e.subnavIsExpanded}],attrs:{id:"the-sidebar"}},[n("div",{ref:"sidebarControl",staticClass:"main-nav",class:{"is-hovering":e.isHovering||!1===e.subnavIsExpanded}},[n("div",{staticClass:"top-nav"},e._l(e.titleNavItems,(function(t,a){return n("NavItem",e._b({key:a,attrs:{"has-custom-icon":""},nativeOn:{click:function(t){return e.toggleSubnav(t)}},scopedSlots:e._u([t.iconCustom&&!t.icon?{key:"item-icon",fn:function(){return[n("div",{domProps:{innerHTML:e._s(t.iconCustom)}})]},proxy:!0}:null],null,!0)},"NavItem",t,!1))})),1),n("div",{staticClass:"bottom-nav"},e._l(e.bottomNavItems,(function(t,a){return n("NavItem",e._b({key:a,attrs:{"has-icon":""}},"NavItem",t,!1))})),1)]),e.subnavIsExpanded?n("Subnav",{attrs:{title:e.titleNavItems[0].name,"title-link":e.titleNavItems[0].link,items:e.topNavItems},scopedSlots:e._u([{key:"top",fn:function(){return[n("MeshSelector",{attrs:{items:e.meshList}})]},proxy:!0}],null,!1,546986014)}):e._e()],1)},u=[],c=(n("7db0"),n("b0c0"),n("b64b"),n("9911"),n("a4d3"),n("e01a"),n("d28b"),n("d3b7"),n("3ca3"),n("ddb0"),n("dde1"));function l(e,t){var n;if("undefined"===typeof Symbol||null==e[Symbol.iterator]){if(Array.isArray(e)||(n=Object(c["a"])(e))||t&&e&&"number"===typeof e.length){n&&(e=n);var a=0,s=function(){};return{s:s,n:function(){return a>=e.length?{done:!0}:{done:!1,value:e[a++]}},e:function(e){throw e},f:s}}throw new TypeError("Invalid attempt to iterate non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}var i,r=!0,o=!1;return{s:function(){n=e[Symbol.iterator]()},n:function(){var e=n.next();return r=e.done,e},e:function(e){o=!0,i=e},f:function(){try{r||null==n["return"]||n["return"]()}finally{if(o)throw i}}}}var d=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"nav-item",class:[{"is-active":e.isActive},{"is-menu-item":e.isMenuItem},{"is-disabled":e.isDisabled},{"is-title":e.title},{"is-nested":e.nested}]},[e._t("default"),n("router-link",{attrs:{to:e.routerLink},nativeOn:{click:function(t){return e.onNavItemClick(t)}}},[e.hasIcon||e.hasCustomIcon?n("div",{staticClass:"nav-icon"},[e._t("item-icon",[e.hasIcon&&e.icon?n("KIcon",{attrs:{width:"18",height:"18","view-box":"0 0 18 18",color:"var(--SidebarIconColor)",icon:e.icon}}):e._e()])],2):e._e(),e.title?n("div",{staticClass:"title-text"},[n("span",{staticClass:"text-uppercase"},[e._v(" "+e._s(e.name)+" ")])]):n("div",{staticClass:"nav-link"},[e._t("item-link",[e._v(" "+e._s(e.name)+" ")])],2),e._t("default")],2)],2)},m=[],h=(n("45fc"),n("ac1f"),n("1276"),n("027b")),p=n("75bb"),f={name:"NavItem",props:{link:{type:String,default:"",required:!1},name:{type:String,default:""},icon:{type:String,default:""},hasIcon:{type:Boolean,default:!1},hasCustomIcon:{type:Boolean,default:!1},isMenuItem:{type:Boolean,default:!0},isDisabled:{type:Boolean,default:!1},title:{type:Boolean,default:!1},nested:{type:Boolean,default:!1}},data:function(){return{meshPath:null}},computed:Object(i["a"])(Object(i["a"])({},Object(r["c"])({selectedMesh:"getSelectedMesh"})),{},{routerLink:function(){var e=this,t={mesh:this.selectedMesh},n=function(){return e.link?{name:e.link,params:t}:e.title?{name:null}:{name:e.$route.name,params:t}};return n()},isActive:function(){var e=this.link,t=this.$route,n=this.$route.path.split("/")[2];return e===t.name||(n===this.routerLink.name||e&&t.matched.some((function(t){return e===t.name||e===t.redirect})))}}),methods:{onNavItemClick:function(){h["a"].logger.info(p["a"].SIDEBAR_ITEM_CLICKED,{data:this.routerLink})}}},v=f,b=(n("2467"),n("55b3"),n("2877")),g=Object(b["a"])(v,d,m,!1,null,"6420e39a",null),_=g.exports,C=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"secondary-nav",class:{"is-collapsed":e.isCollapsed}},[n("div",{staticClass:"mt-3"},[e._t("top")],2),n("div",{staticClass:"subnav-title"},[n("span",{staticClass:"text-uppercase"},[e._t("title",[n("router-link",{attrs:{to:{name:e.titleLink}}},[e._v(" "+e._s(e.title)+" ")])])],2)]),e._t("bottom"),e._l(e.items,(function(t,a){return n("NavItem",e._b({key:a,attrs:{nested:t.nested}},"NavItem",t,!1))}))],2)},I=[],y={name:"SubNav",components:{NavItem:_},props:{title:{type:String,default:""},items:{type:Array,required:!0},titleLink:{type:String,default:""}},data:function(){return{isCollapsed:!1}},computed:{touchDevice:function(){return!(!("ontouchstart"in window)&&!navigator.maxTouchPoints)}},methods:{handleToggle:function(){this.touchDevice&&(this.isCollapsed=!this.isCollapsed,this.$emit("toggled",this.isCollapsed))}}},k=y,x=(n("f561"),n("42e7"),Object(b["a"])(k,C,I,!1,null,"fd5c8ef6",null)),S=x.exports,E=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"mesh-selector-container px-4 pb-4"},[e.items?n("div",[n("h3",{staticClass:"menu-title uppercase"},[e._v(" Filter by Mesh: ")]),n("select",{staticClass:"mesh-selector",attrs:{id:"mesh-selector",name:"mesh-selector"},on:{change:e.changeMesh}},[n("option",{attrs:{value:"all"},domProps:{selected:"all"===e.selectedMesh}},[e._v(" All Meshes ")]),e._l(e.items.items,(function(t){return n("option",{key:t.name,domProps:{value:t.name,selected:t.name===e.selectedMesh}},[e._v(" "+e._s(t.name)+" ")])}))],2)]):n("KAlert",{attrs:{appearance:"danger","alert-message":"No meshes found!"}})],1)},O=[],N={name:"MeshSelector",props:{items:{type:Object,required:!0}},computed:Object(i["a"])({},Object(r["c"])({selectedMesh:"getSelectedMesh"})),methods:{changeMesh:function(e){var t=e.target.value;this.$store.dispatch("updateSelectedMesh",t),localStorage.setItem("selectedMesh",t),this.$root.$router.push({params:{mesh:t}}).catch((function(){}))}}},M=N,j=(n("3372"),Object(b["a"])(M,E,O,!1,null,"86c508b8",null)),T=j.exports,$=n("c6ec"),A={name:"Sidebar",components:{MeshSelector:T,NavItem:_,Subnav:S},data:function(){return{isCollapsed:!1,sidebarSavedState:null,toggleWorkspaces:!1,isHovering:!1,subnavIsExpanded:!0}},computed:Object(i["a"])(Object(i["a"])({},Object(r["d"])("sidebar",{menu:function(e){return e.menu}})),{},{titleNavItems:function(){return this.menu.find((function(e){return"top"===e.position})).items},topNavItems:function(){return this.menu.find((function(e){return"top"===e.position})).items[0].subNav.items},bottomNavItems:function(){return this.menu.find((function(e){return"bottom"===e.position})).items},hasSubnav:function(){var e,t,n;return Boolean(null===(e=this.selectedMenuItem)||void 0===e||null===(t=e.subNav)||void 0===t||null===(n=t.items)||void 0===n?void 0:n.length)},lastMenuList:function(){return Object.keys(this.menuList.sections).length-1},meshList:function(){return this.$store.state.meshes},selectedMenuItem:function(){var e,t=this.$route,n=l(this.menu);try{for(n.s();!(e=n.n()).done;){var a,s=e.value,i=l(s.items);try{for(i.s();!(a=i.n()).done;){var r=a.value,o=t.name!==r.link,u=o&&!t.meta.hideSubnav;if(u)return r}}catch(c){i.e(c)}finally{i.f()}}}catch(c){n.e(c)}finally{n.f()}return null},touchDevice:function(){return!(!("ontouchstart"in window)&&!navigator.maxTouchPoints)}}),mounted:function(){this.sidebarEvent()},beforeDestroy:function(){},methods:{getNavItems:function(e,t,n){return e.find((function(e){return e.position===t})).items},handleResize:function(){var e=$["a"].innerWidth;e<=900&&(this.isCollapsed=!0,this.subnavIsExpanded=!1,this.isHovering=!1),e>=900&&(this.isCollapsed=!1,this.isHovering=!0)},toggleSubnav:function(){this.subnavIsExpanded=!0,this.isCollapsed=!0,localStorage.setItem("sidebarCollapsed",this.subnavIsExpanded)},sidebarEvent:function(){var e=this,t=this.touchDevice,n=this.$refs.sidebarControl;this.$route.params.expandSidebar&&!0===this.$route.params.expandSidebar&&(this.subnavIsExpanded=!0,localStorage.setItem("sidebarCollapsed",!0)),t?(n.addEventListener("touchstart",(function(){e.isHovering=!0})),n.addEventListener("touchend",(function(){e.isHovering=!1}))):(n.addEventListener("mouseover",(function(){e.isHovering=!0})),n.addEventListener("mouseout",(function(){e.isHovering=!1})),n.addEventListener("click",(function(){e.isHovering=!1})))}}},L=A,R=(n("7148"),Object(b["a"])(L,o,u,!1,null,null,null)),w=R.exports,B=function(){var e=this,t=e.$createElement,n=e._self._c||t;return!1===e.alertClosed?n("div",{staticClass:"onboarding-check"},[n("KAlert",{staticClass:"dismissible",attrs:{appearance:"info","is-dismissible":""},on:{closed:e.closeAlert},scopedSlots:e._u([{key:"alertMessage",fn:function(){return[n("div",{staticClass:"alert-content"},[n("div",[n("strong",[e._v("Welcome to "+e._s(e.productName)+"!")]),e._v(" We've detected that you don't have any data plane proxies running yet. We've created an onboarding process to help you! ")]),n("div",[n("KButton",{staticClass:"action-button",attrs:{appearance:"primary",size:"small",to:{path:"/get-started"}}},[e._v(" Get Started ")])],1)])]},proxy:!0}],null,!1,3584331343)})],1):e._e()},D=[],H={name:"OnboardingCheck",data:function(){return{alertClosed:!1,productName:$["g"]}},methods:{closeAlert:function(){this.alertClosed=!0}}},K=H,P=(n("0b00"),Object(b["a"])(K,B,D,!1,null,"0741a11a",null)),q=P.exports,W=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",[e.hideBreadcrumbs?e._e():n("Krumbs",{attrs:{items:e.routes}})],1)},G=[],U=(n("4de4"),n("c975"),n("d81d"),n("07ac"),n("498a"),n("c9e9")),z=n("bc1e"),F={name:"Breadcrumbs",computed:{pageMesh:function(){return this.$route.params.mesh},routes:function(){var e=this,t=[];this.$route.matched.map((function(n){var a=void 0!==n.redirect&&void 0!==n.redirect.name?n.redirect.name:n.name;e.isCurrentRoute(n)&&e.pageMesh&&t.push({key:e.pageMesh,to:{path:"/meshes/".concat(e.pageMesh)},title:"Mesh Overview for ".concat(e.pageMesh),text:e.pageMesh}),e.isCurrentRoute(n)&&n.meta.parent&&"undefined"!==n.meta.parent?t.push({key:n.meta.parent,to:{name:n.meta.parent},title:n.meta.title,text:n.meta.breadcrumb||n.meta.title}):e.isCurrentRoute(n)&&!n.meta.excludeAsBreadcrumb?t.push({key:a,to:{name:a},title:n.meta.title,text:n.meta.breadcrumb||n.meta.title}):n.meta.parent&&"undefined"!==n.meta.parent&&t.push({key:n.meta.parent,to:{name:n.meta.parent},title:n.meta.title,text:n.meta.breadcrumb||n.meta.title})}));var n=this.calculateRouteTextAdvanced(this.$route);return n&&t.push({title:n,text:n}),t},hideBreadcrumbs:function(){return this.$route.query.hide_breadcrumb}},methods:{getBreadcrumbItem:function(e,t,n,a){return{key:e,to:t,title:n,text:a}},isCurrentRoute:function(e){return e.name&&e.name===this.$router.currentRoute.name||e.redirect===this.$router.currentRoute.name},calculateRouteFromQuery:function(e){var t=e.entity_id,n=e.entity_type;if(t&&n){var a=this.$router.resolve({name:"show-".concat(n.split("_")[0]),params:{id:t.split(",")[0]}}).normalizedTo,s=Object(i["a"])(Object(i["a"])({},a),{},{meta:Object(i["a"])({},a.meta)}),r=s.params.id.split("-")[0];return t.split(",").length>1&&t.split(",")[1]&&(r=t.split(",")[1]),s.meta.breadcrumb=r,[Object(i["a"])({},this.getBreadcrumbItem(s.name,s,this.calculateRouteTitle(s),this.calculateRouteText(s)))]}},calculateRouteText:function(e){if(e.path&&e.path.indexOf(":mesh")>-1){var t=this.$router.currentRoute.params;return(t&&t.mesh&&Object(z["g"])(t.mesh)?t.mesh.split("-")[0].trim():t.mesh)||e.meta.breadcrumb||e.meta.title}return e.meta&&(e.meta.breadcrumb||e.meta.title)||e.name||e.meta.breadcrumb||e.meta.title},calculateRouteTitle:function(e){return e.params&&e.params.mesh||e.path.indexOf(":mesh")>-1&&this.$router.currentRoute.params&&this.$router.currentRoute.params.mesh},calculateRouteTextAdvanced:function(e){var t=e.params,n=(t.expandSidebar,Object(U["a"])(t,["expandSidebar"])),a="mesh-overview"===e.name,s=Object.assign({},n,{mesh:null});return a?t.mesh:Object.values(s).filter((function(e){return e}))[0]}}},J=F,X=(n("e7ab"),Object(b["a"])(J,W,G,!1,null,null,null)),Q=X.exports,V={name:"Shell",components:{Breadcrumbs:Q,Sidebar:w,OnboardingCheck:q},computed:Object(i["a"])({},Object(r["c"])({showOnboarding:"showOnboarding"}))},Y=V,Z=Object(b["a"])(Y,a,s,!1,null,null,null);t["default"]=Z.exports},e7ab:function(e,t,n){"use strict";n("5f76")},f122:function(e,t,n){},f561:function(e,t,n){"use strict";n("f122")}}]);