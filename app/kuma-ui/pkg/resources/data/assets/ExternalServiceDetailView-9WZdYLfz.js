import{_ as E}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-DNjspper.js";import{d as w,a as i,o as r,b as n,w as e,e as o,E as V,z as R,c as S,m as p,W as h,f as c,t as m,T as B,q as f}from"./index-CvRMgvyl.js";import{T as $}from"./TagList-D3FCMK7L.js";import"./CodeBlock-DAAzxM62.js";import"./toYaml-DB9FPXFY.js";const b={key:2,class:"stack"},D={class:"columns"},W=w({__name:"ExternalServiceDetailView",setup(F){return(T,M)=>{const x=i("KCard"),_=i("DataSource"),v=i("AppView"),C=i("RouteView");return r(),n(C,{name:"external-service-detail-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:s,t:l})=>[o(v,null,{default:e(()=>[o(_,{src:`/meshes/${s.params.mesh}/external-services/${s.params.service}`},{default:e(({data:t,error:u})=>[u?(r(),n(V,{key:0,error:u},null,8,["error"])):t===void 0?(r(),n(R,{key:1})):(r(),S("div",b,[o(x,{"data-testid":"external-service-details"},{default:e(()=>[p("div",D,[o(h,null,{title:e(()=>[c(m(l("http.api.property.address")),1)]),body:e(()=>[o(B,{text:t.networking.address},null,8,["text"])]),_:2},1024),c(),t.tags?(r(),n(h,{key:0},{title:e(()=>[c(m(l("http.api.property.tags")),1)]),body:e(()=>[o($,{tags:t.tags},null,8,["tags"])]),_:2},1024)):f("",!0)])]),_:2},1024),c(),p("div",null,[p("h3",null,m(l("external-services.detail.config")),1),c(),o(E,{class:"mt-4","data-testid":"external-service-config",resource:t.config,"is-searchable":"",query:s.params.codeSearch,"is-filter-mode":s.params.codeFilter,"is-reg-exp-mode":s.params.codeRegExp,onQueryChange:a=>s.update({codeSearch:a}),onFilterModeChange:a=>s.update({codeFilter:a}),onRegExpModeChange:a=>s.update({codeRegExp:a})},{default:e(({copy:a,copying:y})=>[y?(r(),n(_,{key:0,src:`/meshes/${t.mesh}/external-services/${t.name}/as/kubernetes?no-store`,onChange:d=>{a(g=>g(d))},onError:d=>{a((g,k)=>k(d))}},null,8,["src","onChange","onError"])):f("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])]))]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{W as default};
