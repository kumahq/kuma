(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["shell"],{1359:function(t,e,n){"use strict";n("2cdf")},"16d8":function(t,e,n){"use strict";n("7bb9")},"1b4f":function(t,e,n){},"2cdf":function(t,e,n){},"42e7":function(t,e,n){"use strict";n("45cc")},"45cc":function(t,e,n){},"55b3":function(t,e,n){"use strict";n("1b4f")},"5f76":function(t,e,n){},"66b4":function(t,e,n){},7148:function(t,e,n){"use strict";n("66b4")},"7bb9":function(t,e,n){},"857a":function(t,e,n){var a=n("1d80"),s=/"/g;t.exports=function(t,e,n,i){var r=String(a(t)),o="<"+e;return""!==n&&(o+=" "+n+'="'+String(i).replace(s,"&quot;")+'"'),o+">"+r+"</"+e+">"}},"97d1":function(t,e,n){},9911:function(t,e,n){"use strict";var a=n("23e7"),s=n("857a"),i=n("af03");a({target:"String",proto:!0,forced:i("link")},{link:function(t){return s(this,"a","href",t)}})},af03:function(t,e,n){var a=n("d039");t.exports=function(t){return a((function(){var e=""[t]('"');return e!==e.toLowerCase()||e.split('"').length>3}))}},b6f5:function(t,e,n){},be50:function(t,e,n){"use strict";n("97d1")},c6ec:function(t,e,n){"use strict";n.d(e,"c",(function(){return a})),n.d(e,"b",(function(){return s})),n.d(e,"a",(function(){return i}));var a="Kuma",s=12,i=window},ce5e:function(t,e,n){"use strict";n("b6f5")},deb3:function(t,e,n){"use strict";n.r(e);var a=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"main-content-container"},[n("Sidebar"),n("main",{staticClass:"main-content"},[n("div",{staticClass:"page"},[t.showOnboarding?n("OnboardingCheck"):t._e(),n("Breadcrumbs"),n("router-view")],1)])],1)},s=[],i=(n("b0c0"),n("f3f3")),r=n("2f62"),o=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("aside",{staticClass:"has-subnav",class:[{"is-collapsed":t.isCollapsed},{"subnav-expanded":t.subnavIsExpanded}],attrs:{id:"the-sidebar"}},[n("div",{ref:"sidebarControl",staticClass:"main-nav",class:{"is-hovering":t.isHovering||!1===t.subnavIsExpanded}},[n("div",{staticClass:"top-nav"},t._l(t.titleNavItems,(function(e,a){return n("NavItem",t._b({key:a,attrs:{"has-custom-icon":""},nativeOn:{click:function(e){return t.toggleSubnav(e)}}},"NavItem",e,!1),[e.iconCustom&&!e.icon?n("template",{slot:"item-icon"},[n("div",{domProps:{innerHTML:t._s(e.iconCustom)}})]):t._e()],2)})),1),n("div",{staticClass:"bottom-nav"},t._l(t.bottomNavItems,(function(e,a){return n("NavItem",t._b({key:a,attrs:{"has-icon":""}},"NavItem",e,!1))})),1)]),t.subnavIsExpanded?n("Subnav",{attrs:{title:t.titleNavItems[0].name,"title-link":t.titleNavItems[0].link,items:t.topNavItems}},[n("template",{slot:"top"},[n("MeshSelector",{attrs:{items:t.meshList}})],1)],2):t._e()],1)},u=[],c=(n("7db0"),n("b64b"),n("9911"),n("a4d3"),n("e01a"),n("d28b"),n("d3b7"),n("3ca3"),n("ddb0"),n("dde1"));function l(t,e){var n;if("undefined"===typeof Symbol||null==t[Symbol.iterator]){if(Array.isArray(t)||(n=Object(c["a"])(t))||e&&t&&"number"===typeof t.length){n&&(t=n);var a=0,s=function(){};return{s:s,n:function(){return a>=t.length?{done:!0}:{done:!1,value:t[a++]}},e:function(t){throw t},f:s}}throw new TypeError("Invalid attempt to iterate non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}var i,r=!0,o=!1;return{s:function(){n=t[Symbol.iterator]()},n:function(){var t=n.next();return r=t.done,t},e:function(t){o=!0,i=t},f:function(){try{r||null==n["return"]||n["return"]()}finally{if(o)throw i}}}}var d=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"nav-item",class:[{"is-active":t.isActive},{"is-menu-item":t.isMenuItem},{"is-disabled":t.isDisabled},{"is-title":t.title},{"is-nested":t.nested}]},[t._t("default"),n("router-link",{attrs:{to:t.routerLink}},[t.hasIcon||t.hasCustomIcon?n("div",{staticClass:"nav-icon"},[t._t("item-icon",[t.hasIcon&&t.icon?n("KIcon",{attrs:{width:"18",height:"18","view-box":"0 0 18 18",color:"var(--SidebarIconColor)",icon:t.icon}}):t._e()])],2):t._e(),t.title?n("div",{staticClass:"title-text"},[n("span",{staticClass:"text-uppercase"},[t._v(" "+t._s(t.name)+" ")])]):n("div",{staticClass:"nav-link"},[t._t("item-link",[t._v(" "+t._s(t.name)+" ")])],2),t._t("default")],2)],2)},m=[],h=(n("99af"),n("45fc"),n("ac1f"),n("1276"),{name:"NavItem",props:{link:{type:String,default:"",required:!1},linkObj:{type:Object,default:function(){return null},required:!1},name:{type:String,default:""},icon:{type:String,default:""},hasIcon:{type:Boolean,default:!1},hasCustomIcon:{type:Boolean,default:!1},isMenuItem:{type:Boolean,default:!0},isDisabled:{type:Boolean,default:!1},title:{type:Boolean,default:!1},nested:{type:Boolean,default:!1}},data:function(){return{meshPath:null}},computed:Object(i["a"])(Object(i["a"])(Object(i["a"])({},Object(r["d"])(["selectedMesh"])),Object(r["c"])({multicluster:"config/getMulticlusterStatus"})),{},{linkPath:function(){var t=this.link;return this.link.pathFlip?t.root?this.preparePath(t.url):"".concat(this.preparePath(t.url),"/").concat(this.meshPath):t.root?this.preparePath(t.url):"/".concat(this.meshPath).concat(this.preparePath(t.url))},routerLink:function(){var t,e,n=this,a=!this.subNav&&Object.keys((null===(t=this.$route)||void 0===t?void 0:t.params)||{}).length>0?null===(e=this.$route)||void 0===e?void 0:e.params:void 0,s=function(){return n.linkObj?n.linkObj:n.link?{name:n.link,params:a}:n.title?{name:null}:{name:n.$route.name,params:a}};return s()},isActive:function(){var t=this.link,e=this.$route,n=this.$route.path.split("/")[2];return t===e.name||(n===this.routerLink.name||t&&e.matched.some((function(e){return t===e.name||t===e.redirect})))}}),watch:{selectedMesh:function(){this.setMeshPath()}},beforeMount:function(){this.setMeshPath()},methods:{preparePath:function(t){return"/"===t[0]?t:"/".concat(t)},setMeshPath:function(){var t=localStorage.getItem("selectedMesh"),e=this.$route.params.mesh;e&&e.length>0?this.meshPath=e:t&&t.length>0&&(this.meshPath=t),this.meshPath=this.$store.getters.getSelectedMesh}}}),p=h,f=(n("ce5e"),n("55b3"),n("2877")),v=Object(f["a"])(p,d,m,!1,null,"6c0cf012",null),b=v.exports,g=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"secondary-nav",class:{"is-collapsed":t.isCollapsed}},[n("div",{staticClass:"mt-3"},[t._t("top")],2),n("div",{staticClass:"subnav-title"},[n("span",{staticClass:"text-uppercase"},[t._t("title",[n("router-link",{attrs:{to:{name:t.titleLink}}},[t._v(" "+t._s(t.title)+" ")])])],2)]),t._t("bottom"),t._l(t.items,(function(e,a){return n("NavItem",t._b({key:a,attrs:{nested:e.nested}},"NavItem",e,!1))}))],2)},C=[],y={name:"SubNav",components:{NavItem:b},props:{title:{type:String,default:""},items:{type:Array,required:!0},titleLink:{type:String,default:""}},data:function(){return{isCollapsed:!1}},computed:{touchDevice:function(){return!(!("ontouchstart"in window)&&!navigator.maxTouchPoints)}},methods:{handleToggle:function(){this.touchDevice&&(this.isCollapsed=!this.isCollapsed,this.$emit("toggled",this.isCollapsed))}}},_=y,I=(n("1359"),n("42e7"),Object(f["a"])(_,g,C,!1,null,"214ea3ee",null)),k=I.exports,x=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"mesh-selector-container px-4 pb-4"},[t.items?n("div",[n("h3",{staticClass:"menu-title uppercase"},[t._v(" Filter by Mesh: ")]),n("select",{staticClass:"mesh-selector",attrs:{id:"mesh-selector",name:"mesh-selector"},on:{change:t.changeMesh}},[n("option",{attrs:{value:"all"},domProps:{selected:"all"===t.selectedMesh}},[t._v(" All Meshes ")]),t._l(t.items.items,(function(e){return n("option",{key:e.name,domProps:{value:e.name,selected:e.name===t.selectedMesh}},[t._v(" "+t._s(e.name)+" ")])}))],2)]):n("KAlert",{attrs:{appearance:"danger","alert-message":"No meshes found!"}})],1)},O=[],S={name:"MeshSelector",props:{items:{type:Object,required:!0}},computed:{selectedMesh:function(){var t=localStorage.getItem("selectedMesh"),e=this.$route.params.mesh;return t||e}},methods:{changeMesh:function(t){var e=t.target.value;this.$store.dispatch("updateSelectedMesh",e),localStorage.setItem("selectedMesh",e),this.$root.$router.push({params:{mesh:e}}).catch((function(){}))}}},M=S,j=(n("16d8"),Object(f["a"])(M,x,O,!1,null,"609def6e",null)),$=j.exports,w=n("c6ec"),N={name:"Sidebar",components:{MeshSelector:$,NavItem:b,Subnav:k},data:function(){return{isCollapsed:!1,sidebarSavedState:null,toggleWorkspaces:!1,isHovering:!1,subnavIsExpanded:!0}},computed:Object(i["a"])(Object(i["a"])({},Object(r["d"])("sidebar",{menu:function(t){return t.menu}})),{},{titleNavItems:function(){return this.menu.find((function(t){return"top"===t.position})).items},topNavItems:function(){return this.menu.find((function(t){return"top"===t.position})).items[0].subNav.items},bottomNavItems:function(){return this.menu.find((function(t){return"bottom"===t.position})).items},hasSubnav:function(){var t,e,n;return Boolean(null===(t=this.selectedMenuItem)||void 0===t||null===(e=t.subNav)||void 0===e||null===(n=e.items)||void 0===n?void 0:n.length)},lastMenuList:function(){return Object.keys(this.menuList.sections).length-1},meshList:function(){return this.$store.state.meshes},selectedMenuItem:function(){var t,e=this.$route,n=l(this.menu);try{for(n.s();!(t=n.n()).done;){var a,s=t.value,i=l(s.items);try{for(i.s();!(a=i.n()).done;){var r=a.value,o=e.name!==r.link,u=o&&!e.meta.hideSubnav;if(u)return r}}catch(c){i.e(c)}finally{i.f()}}}catch(c){n.e(c)}finally{n.f()}return null},touchDevice:function(){return!(!("ontouchstart"in window)&&!navigator.maxTouchPoints)}}),mounted:function(){this.sidebarEvent()},beforeDestroy:function(){},methods:{getNavItems:function(t,e,n){return t.find((function(t){return t.position===e})).items},handleResize:function(){var t=w["a"].innerWidth;t<=900&&(this.isCollapsed=!0,this.subnavIsExpanded=!1,this.isHovering=!1),t>=900&&(this.isCollapsed=!1,this.isHovering=!0)},toggleSubnav:function(){this.subnavIsExpanded=!0,this.isCollapsed=!0,localStorage.setItem("sidebarCollapsed",this.subnavIsExpanded)},sidebarEvent:function(){var t=this,e=this.touchDevice,n=this.$refs.sidebarControl;this.$route.params.expandSidebar&&!0===this.$route.params.expandSidebar&&(this.subnavIsExpanded=!0,localStorage.setItem("sidebarCollapsed",!0)),e?(n.addEventListener("touchstart",(function(){t.isHovering=!0})),n.addEventListener("touchend",(function(){t.isHovering=!1}))):(n.addEventListener("mouseover",(function(){t.isHovering=!0})),n.addEventListener("mouseout",(function(){t.isHovering=!1})),n.addEventListener("click",(function(){t.isHovering=!1})))}}},E=N,P=(n("7148"),Object(f["a"])(E,o,u,!1,null,null,null)),R=P.exports,B=function(){var t=this,e=t.$createElement,n=t._self._c||e;return!1===t.alertClosed?n("div",{staticClass:"onboarding-check"},[n("KAlert",{staticClass:"dismissible",attrs:{appearance:"info","is-dismissible":""},on:{closed:t.closeAlert}},[n("template",{slot:"alertMessage"},[n("div",{staticClass:"alert-content"},[n("div",[n("strong",[t._v("Welcome to "+t._s(t.productName)+"!")]),t._v(" We've detected that you don't have any data plane proxies running yet. We've created an onboarding process to help you! ")]),n("div",[n("KButton",{staticClass:"action-button",attrs:{appearance:"primary",size:"small",to:{path:"/get-started"}}},[t._v(" Get Started ")])],1)])])],2)],1):t._e()},L=[],A={name:"OnboardingCheck",data:function(){return{alertClosed:!1,productName:w["c"]}},methods:{closeAlert:function(){this.alertClosed=!0}}},T=A,H=(n("be50"),Object(f["a"])(T,B,L,!1,null,"bf1fa6c6",null)),D=H.exports,q=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",[t.hideBreadcrumbs?t._e():n("Krumbs",{attrs:{items:t.routes}})],1)},K=[],W=(n("4de4"),n("c975"),n("d81d"),n("07ac"),n("498a"),n("c9e9")),z=n("bc1e"),F={name:"Breadcrumbs",computed:{pageMesh:function(){return this.$route.params.mesh},routes:function(){var t=this,e=[];this.$route.matched.map((function(n){var a=void 0!==n.redirect&&void 0!==n.redirect.name?n.redirect.name:n.name;t.isCurrentRoute(n)&&t.pageMesh&&e.push({key:t.pageMesh,to:{path:"/meshes/".concat(t.pageMesh)},title:"Mesh Overview for ".concat(t.pageMesh),text:t.pageMesh}),t.isCurrentRoute(n)&&n.meta.parent&&"undefined"!==n.meta.parent?e.push({key:n.meta.parent,to:{name:n.meta.parent},title:n.meta.title,text:n.meta.breadcrumb||n.meta.title}):t.isCurrentRoute(n)&&!n.meta.excludeAsBreadcrumb?e.push({key:a,to:{name:a},title:n.meta.title,text:n.meta.breadcrumb||n.meta.title}):n.meta.parent&&"undefined"!==n.meta.parent&&e.push({key:n.meta.parent,to:{name:n.meta.parent},title:n.meta.title,text:n.meta.breadcrumb||n.meta.title})}));var n=this.calculateRouteTextAdvanced(this.$route);return n&&e.push({title:n,text:n}),e},hideBreadcrumbs:function(){return this.$route.query.hide_breadcrumb}},methods:{getBreadcrumbItem:function(t,e,n,a){return{key:t,to:e,title:n,text:a}},isCurrentRoute:function(t){return t.name&&t.name===this.$router.currentRoute.name||t.redirect===this.$router.currentRoute.name},calculateRouteFromQuery:function(t){var e=t.entity_id,n=t.entity_type;if(e&&n){var a=this.$router.resolve({name:"show-".concat(n.split("_")[0]),params:{id:e.split(",")[0]}}).normalizedTo,s=Object(i["a"])(Object(i["a"])({},a),{},{meta:Object(i["a"])({},a.meta)}),r=s.params.id.split("-")[0];return e.split(",").length>1&&e.split(",")[1]&&(r=e.split(",")[1]),s.meta.breadcrumb=r,[Object(i["a"])({},this.getBreadcrumbItem(s.name,s,this.calculateRouteTitle(s),this.calculateRouteText(s)))]}},calculateRouteText:function(t){if(t.path&&t.path.indexOf(":mesh")>-1){var e=this.$router.currentRoute.params;return(e&&e.mesh&&Object(z["h"])(e.mesh)?e.mesh.split("-")[0].trim():e.mesh)||t.meta.breadcrumb||t.meta.title}return t.meta&&(t.meta.breadcrumb||t.meta.title)||t.name||t.meta.breadcrumb||t.meta.title},calculateRouteTitle:function(t){return t.params&&t.params.mesh||t.path.indexOf(":mesh")>-1&&this.$router.currentRoute.params&&this.$router.currentRoute.params.mesh},calculateRouteTextAdvanced:function(t){var e=t.params,n=(e.expandSidebar,Object(W["a"])(e,["expandSidebar"])),a="mesh-overview"===t.name,s=Object.assign({},n,{mesh:null});return a?e.mesh:Object.values(s).filter((function(t){return t}))[0]}}},J=F,G=(n("e7ab"),Object(f["a"])(J,q,K,!1,null,null,null)),Q=G.exports,U={name:"Shell",components:{Breadcrumbs:Q,Sidebar:R,OnboardingCheck:D},computed:Object(i["a"])(Object(i["a"])({},Object(r["d"])({dpCount:"totalDataplaneCount",meshes:"meshes"})),{},{showOnboarding:function(){var t=1===this.meshes.total&&"default"===this.meshes.items[0].name,e=0===this.dpCount;return e&&t}})},V=U,X=Object(f["a"])(V,a,s,!1,null,null,null);e["default"]=X.exports},e7ab:function(t,e,n){"use strict";n("5f76")}}]);