import{d as M,e as o,o as r,m as d,w as a,a as t,k as p,X as _,b as s,c as h,H as f,J as g,t as C,p as R,q as B}from"./index-Yqc5mH7h.js";import{_ as D}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-7LOb7PdN.js";const F={class:"stack"},K={class:"columns"},$=M({__name:"MeshMultiZoneServiceDetailView",props:{data:{}},setup(v){const c=v;return(x,N)=>{const S=o("KumaPort"),u=o("KTruncate"),V=o("XBadge"),w=o("KCard"),b=o("DataSource"),k=o("AppView"),y=o("RouteView");return r(),d(y,{name:"mesh-multi-zone-service-detail-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:n})=>[t(k,null,{default:a(()=>[p("div",F,[t(w,null,{default:a(()=>[p("div",K,[t(_,null,{title:a(()=>[s(`
                Ports
              `)]),body:a(()=>[t(u,null,{default:a(()=>[(r(!0),h(f,null,g(c.data.spec.ports,e=>(r(),d(S,{key:e.port,port:{...e,targetPort:void 0}},null,8,["port"]))),128))]),_:1})]),_:1}),s(),t(_,null,{title:a(()=>[s(`
                Selector
              `)]),body:a(()=>[t(u,null,{default:a(()=>[(r(!0),h(f,null,g(x.data.spec.selector.meshService.matchLabels,(e,l)=>(r(),d(V,{key:`${l}:${e}`,appearance:"info"},{default:a(()=>[s(C(l)+":"+C(e),1)]),_:2},1024))),128))]),_:1})]),_:1})])]),_:1}),s(),p("div",null,[t(D,{resource:c.data.config,"is-searchable":"",query:n.params.codeSearch,"is-filter-mode":n.params.codeFilter,"is-reg-exp-mode":n.params.codeRegExp,onQueryChange:e=>n.update({codeSearch:e}),onFilterModeChange:e=>n.update({codeFilter:e}),onRegExpModeChange:e=>n.update({codeRegExp:e})},{default:a(({copy:e,copying:l})=>[l?(r(),d(b,{key:0,src:`/meshes/${c.data.mesh}/mesh-multi-zone-service/${c.data.id}/as/kubernetes?no-store`,onChange:i=>{e(m=>m(i))},onError:i=>{e((m,E)=>E(i))}},null,8,["src","onChange","onError"])):R("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])]),_:2},1024)]),_:1})}}}),T=B($,[["__scopeId","data-v-b51f7db2"]]);export{T as default};
