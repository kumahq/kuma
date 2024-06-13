import{_ as V}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BJHqPy6s.js";import{d as D,i as n,o as i,a as y,w as a,j as t,P as E,k as c,b as _,g as m,t as g,H as L,J as S,e as C,Y as F}from"./index-DKbsM-FP.js";import"./CodeBlock-DMM9-fqU.js";import"./toYaml-DB9FPXFY.js";const B={key:0,class:"stack","data-testid":"detail-view-details"},I={class:"mt-4"},K={"data-testid":"affected-data-plane-proxies"},H=D({__name:"PolicyDetailView",setup(M){return(N,T)=>{const x=n("RouteTitle"),w=n("KInput"),k=n("RouterLink"),h=n("DataCollection"),u=n("DataLoader"),v=n("KCard"),f=n("DataSource"),R=n("AppView"),$=n("RouteView");return i(),y($,{name:"policy-detail-view",params:{mesh:"",policy:"",policyPath:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,dataplane:""}},{default:a(({route:e,t:r})=>[t(f,{src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}`},{default:a(({data:p,error:b})=>[t(R,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"policy-list-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath}},text:r("policies.routes.item.breadcrumbs")}]},E({default:a(()=>[c(),t(u,{data:[p],errors:[b]},{default:a(()=>[p?(i(),_("div",B,[t(v,null,{default:a(()=>[m("h2",null,g(r("policies.detail.affected_dpps")),1),c(),t(u,{src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/dataplanes`},{default:a(({data:s})=>[m("div",I,[t(h,{items:s.items},{default:a(({items:d})=>[t(w,{"model-value":e.params.dataplane,type:"search",placeholder:r("policies.detail.dataplane_input_placeholder"),"data-testid":"dataplane-search-input",onInput:o=>e.update({dataplane:o})},null,8,["model-value","placeholder","onInput"]),c(),t(h,{items:d,predicate:o=>o.name.toLowerCase().includes(e.params.dataplane.toLowerCase()),empty:!1},{default:a(({items:o})=>[m("ul",K,[(i(!0),_(L,null,S(o,l=>(i(),_("li",{key:l.name,"data-testid":"dataplane-name"},[t(k,{to:{name:"data-plane-detail-view",params:{mesh:l.mesh,dataPlane:l.name}}},{default:a(()=>[c(g(l.name),1)]),_:2},1032,["to"])]))),128))])]),_:2},1032,["items","predicate"])]),_:2},1032,["items"])])]),_:2},1032,["src"])]),_:2},1024),c(),t(V,{resource:p.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:s=>e.update({codeSearch:s}),onFilterModeChange:s=>e.update({codeFilter:s}),onRegExpModeChange:s=>e.update({codeRegExp:s})},{default:a(({copy:s,copying:d})=>[d?(i(),y(f,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:o=>{s(l=>l(o))},onError:o=>{s((l,P)=>P(o))}},null,8,["src","onChange","onError"])):C("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])):C("",!0)]),_:2},1032,["data","errors"])]),_:2},[p?{name:"title",fn:a(()=>[m("h1",null,[t(F,{text:p.name},{default:a(()=>[t(x,{title:r("policies.routes.item.title",{name:p.name})},null,8,["title"])]),_:2},1032,["text"])])]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{H as default};
