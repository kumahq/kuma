import{d as R,r as n,o as d,p as m,w as a,b as o,l as p,m as f,ar as h,Q as x,e as i,t as u,q as C}from"./index-BIN9nSPF.js";import{T as S}from"./TagList-P3Qih_Hg.js";import{_ as B}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-3fFCInp0.js";const F={class:"stack"},q=R({__name:"ExternalServiceDetailView",setup(T){return(A,s)=>{const v=n("XCopyButton"),y=n("XAboutCard"),b=n("DataSource"),w=n("DataLoader"),V=n("AppView"),E=n("RouteView");return d(),m(E,{name:"external-service-detail-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:l,uri:g})=>[o(V,null,{default:a(()=>[p("div",F,[o(w,{src:g(f(h),"/meshes/:mesh/external-services/:name",{mesh:e.params.mesh,name:e.params.service})},{default:a(({data:r})=>[o(y,{"data-testid":"external-service-details",title:l("external-services.detail.about.title"),created:r.creationTime,modified:r.modificationTime},{default:a(()=>[o(x,{layout:"horizontal"},{title:a(()=>[i(u(l("http.api.property.address")),1)]),body:a(()=>[o(v,{variant:"badge",format:"default",text:r.networking.address},null,8,["text"])]),_:2},1024),s[2]||(s[2]=i()),r.tags?(d(),m(x,{key:0,layout:"horizontal"},{title:a(()=>[i(u(l("http.api.property.tags")),1)]),body:a(()=>[o(S,{tags:r.tags},null,8,["tags"])]),_:2},1024)):C("",!0)]),_:2},1032,["title","created","modified"]),s[4]||(s[4]=i()),p("div",null,[p("h3",null,u(l("external-services.detail.config")),1),s[3]||(s[3]=i()),o(B,{class:"mt-4","data-testid":"external-service-config",resource:r.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:t=>e.update({codeSearch:t}),onFilterModeChange:t=>e.update({codeFilter:t}),onRegExpModeChange:t=>e.update({codeRegExp:t})},{default:a(({copy:t,copying:k})=>[k?(d(),m(b,{key:0,src:g(f(h),"/meshes/:mesh/external-services/:name/as/kubernetes",{mesh:e.params.mesh,name:e.params.service},{cacheControl:"no-store"}),onChange:c=>{t(_=>_(c))},onError:c=>{t((_,D)=>D(c))}},null,8,["src","onChange","onError"])):C("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});export{q as default};
