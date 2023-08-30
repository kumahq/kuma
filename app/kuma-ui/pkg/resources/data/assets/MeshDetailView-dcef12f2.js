import{h as y,f as $,l as B,m as A,A as R,i as k,n as S,E as w,o as N,_ as b}from"./RouteView.vue_vue_type_script_setup_true_lang-07b4cab0.js";import{_ as C}from"./RouteTitle.vue_vue_type_script_setup_true_lang-9295e5a4.js";import{_ as I}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-0d975278.js";import{d as g,c as D,o as a,e as x,g as o,t as m,b as s,h as e,m as L,a as i,w as d,k as E}from"./index-5cc1bd9e.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-bebc6d7a.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-c839ae76.js";const K={class:"date-status"},j=g({__name:"ResourceDateStatus",props:{creationTime:{},modificationTime:{}},setup(l){const c=l,{t:r,formatIsoDate:n}=y(),u=D(()=>n(c.creationTime)),p=D(()=>n(c.modificationTime));return(_,t)=>(a(),x("span",K,[o(m(s(r)("common.detail.created"))+": "+m(u.value)+" ",1),e(s(L),{icon:"arrowRight"}),o(" "+m(s(r)("common.detail.modified"))+": "+m(p.value),1)]))}});const q=$(j,[["__scopeId","data-v-5d50f5d4"]]),z={key:4,class:"stack","data-testid":"detail-view-details"},F={class:"date-status-wrapper"},G=g({__name:"MeshDetailView",setup(l){const{t:c}=y(),r=B(),n=A();return(u,p)=>(a(),i(b,{name:"mesh-overview-view"},{default:d(({route:_})=>[e(C,{title:s(c)("meshes.routes.overview.title")},null,8,["title"]),o(),e(R,null,{default:d(()=>[e(k,{src:`/meshes/${_.params.mesh}`},{default:d(({data:t,isLoading:T,error:f})=>[e(k,{src:`/mesh-insights/${_.params.mesh}`},{default:d(({data:h,isLoading:V,error:v})=>[T||V?(a(),i(S,{key:0})):f?(a(),i(w,{key:1,error:f},null,8,["error"])):v?(a(),i(w,{key:2,error:v},null,8,["error"])):t===void 0||h===void 0?(a(),i(N,{key:3})):(a(),x("div",z,[e(s(n),{mesh:t,"mesh-insight":h},null,8,["mesh","mesh-insight"]),o(),e(I,{id:"code-block-mesh",resource:t,"resource-fetcher":M=>s(r).getMesh({name:_.params.mesh},M)},null,8,["resource","resource-fetcher"]),o(),E("div",F,[e(q,{"creation-time":t.creationTime,"modification-time":t.modificationTime},null,8,["creation-time","modification-time"])])]))]),_:2},1032,["src"])]),_:2},1032,["src"])]),_:2},1024)]),_:1}))}});const W=$(G,[["__scopeId","data-v-ac1a8d22"]]);export{W as default};
