import{d as R,r,o,p as m,w as t,b as i,Q as y,e as s,t as c,c as u,J as _,K as C,l as B,q as D,_ as F}from"./index-gI7YoWPY.js";import{_ as X}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CfFLyldn.js";const z=R({__name:"MeshMultiZoneServiceDetailView",props:{data:{}},setup(v){const a=v;return(f,p)=>{const k=r("KumaPort"),b=r("XBadge"),x=r("XAboutCard"),S=r("DataSource"),V=r("XLayout"),w=r("AppView"),E=r("RouteView");return o(),m(E,{name:"mesh-multi-zone-service-detail-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:n,t:d})=>[i(w,null,{default:t(()=>[i(V,{type:"stack"},{default:t(()=>[i(x,{title:d("services.mesh-multi-zone-service.about.title"),created:a.data.creationTime,modified:a.data.modificationTime},{default:t(()=>[i(y,{layout:"horizontal"},{title:t(()=>[s(c(d("http.api.property.ports")),1)]),body:t(()=>[a.data.spec.ports.length?(o(!0),u(_,{key:0},C(a.data.spec.ports,e=>(o(),m(k,{key:e.port,port:{...e,targetPort:void 0}},null,8,["port"]))),128)):(o(),u(_,{key:1},[s(c(d("common.detail.none")),1)],64))]),_:2},1024),p[2]||(p[2]=s()),i(y,{layout:"horizontal"},{title:t(()=>[s(c(d("http.api.property.selector")),1)]),body:t(()=>[Object.keys(f.data.spec.selector.meshService.matchLabels).length?(o(!0),u(_,{key:0},C(f.data.spec.selector.meshService.matchLabels,(e,l)=>(o(),m(b,{key:`${l}:${e}`,appearance:"info"},{default:t(()=>[s(c(l)+":"+c(e),1)]),_:2},1024))),128)):(o(),u(_,{key:1},[s(c(d("common.detail.none")),1)],64))]),_:2},1024)]),_:2},1032,["title","created","modified"]),p[3]||(p[3]=s()),B("div",null,[i(X,{resource:a.data.config,"is-searchable":"",query:n.params.codeSearch,"is-filter-mode":n.params.codeFilter,"is-reg-exp-mode":n.params.codeRegExp,onQueryChange:e=>n.update({codeSearch:e}),onFilterModeChange:e=>n.update({codeFilter:e}),onRegExpModeChange:e=>n.update({codeRegExp:e})},{default:t(({copy:e,copying:l})=>[l?(o(),m(S,{key:0,src:`/meshes/${a.data.mesh}/mesh-multi-zone-service/${a.data.id}/as/kubernetes?no-store`,onChange:h=>{e(g=>g(h))},onError:h=>{e((g,M)=>M(h))}},null,8,["src","onChange","onError"])):D("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])]),_:2},1024)]),_:2},1024)]),_:1})}}}),A=F(z,[["__scopeId","data-v-969d10f4"]]);export{A as default};
